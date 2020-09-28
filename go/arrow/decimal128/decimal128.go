// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package decimal128 // import "github.com/apache/arrow/go/arrow/decimal128"
import (
	"fmt"
	"github.com/apache/arrow/go/arrow/internal/debug"
	"strconv"
)

var (
	MaxDecimal128 = New(542101086242752217, 687399551400673280-1)
)

// Num represents a signed 128-bit integer in two's complement.
// Calculations wrap around and overflow is ignored.
//
// For a discussion of the algorithms, look at Knuth's volume 2,
// Semi-numerical Algorithms section 4.3.1.
//
// Adapted from the Apache ORC C++ implementation
type Num struct {
	lo uint64 // low bits
	hi int64  // high bits
}

// New returns a new signed 128-bit integer value.
func New(hi int64, lo uint64) Num {
	return Num{lo: lo, hi: hi}
}

var ScaleMultipliers = []Num{
	{hi:0, lo:1},
	{hi:0, lo:10},
	{hi:0, lo:100},
	{hi:0, lo:1000},
	{hi:0, lo:10000},
	{hi:0, lo:100000},
	{hi:0, lo:1000000},
	{hi:0, lo:10000000},
	{hi:0, lo:100000000},
	{hi:0, lo:1000000000},
	{hi:0, lo:10000000000},
	{hi:0, lo:100000000000},
	{hi:0, lo:1000000000000},
	{hi:0, lo:10000000000000},
	{hi:0, lo:100000000000000},
	{hi:0, lo:1000000000000000},
	{hi:0, lo:10000000000000000},
	{hi:0, lo:100000000000000000},
	{hi:0, lo:1000000000000000000},
	{hi:0, lo:10000000000000000000},
	{hi:5, lo:7766279631452241920},
	{hi:54, lo:3875820019684212736},
	{hi:542, lo:1864712049423024128},
	{hi:5421, lo:200376420520689664},
	{hi:54210, lo:2003764205206896640},
	{hi:542101, lo:1590897978359414784},
	{hi:5421010, lo:15908979783594147840},
	{hi:54210108, lo:11515845246265065472},
	{hi:542101086, lo:4477988020393345024},
	{hi:5421010862, lo:7886392056514347008},
	{hi:54210108624, lo:5076944270305263616},
	{hi:542101086242, lo:13875954555633532928},
	{hi:5421010862427, lo:9632337040368467968},
	{hi:54210108624275, lo:4089650035136921600},
	{hi:542101086242752, lo:4003012203950112768},
	{hi:5421010862427522, lo:3136633892082024448},
	{hi:54210108624275221, lo:uint64(12919594847110692864)},
	{hi:542101086242752217, lo:68739955140067328},
	{hi:5421010862427522170, lo:687399551400673280},
}

var kIntMask = uint64(0xFFFFFFFF)
var kCarryBit = int64(1) << 32;

func GetScaleMultiplier(scale int) Num {
	debug.Assert(scale <= 38, "scale must be less equal than 38")
	debug.Assert(scale >= 0, "scale must be greater equal than 0")
	return ScaleMultipliers[scale]
}

// FromU64 returns a new signed 128-bit integer value from the provided uint64 one.
func FromU64(v uint64) Num {
	return New(0, v)
}

// FromI64 returns a new signed 128-bit integer value from the provided int64 one.
func FromI64(v int64) Num {
	switch {
	case v > 0:
		return New(0, uint64(v))
	case v < 0:
		return New(-1, uint64(v))
	default:
		return Num{}
	}
}

func ShiftAndAdd(data string, out *Num) {
	length := len(data)
	for posn := 0; posn < length; {
		group_size := 18 // int64_max_digits_count, 9223372036854775807
		if group_size > length-posn {
			group_size = length-posn
		}
		chunk, err := strconv.ParseInt(data[posn:posn+group_size],10,64)
		if err != nil {
			panic(fmt.Errorf("arrow/decimal: empty string cannot be converted to decimal"))
		}

		out.Mul(GetScaleMultiplier(group_size))
		out.Add(chunk)
		posn += group_size
	}
}

func StringToInteger(whole_digits, fractional_digits string) *Num {
	debug.Assert(len(whole_digits) + len(fractional_digits) > 0,
		"length of parsed decimal string should be greater than 0")
	out := &Num{}
	ShiftAndAdd(whole_digits, out)
	ShiftAndAdd(fractional_digits, out)
	return out
}

func FromString(input string, precision *int, scale *int) (*Num, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("arrow/decimal: empty string cannot be converted to decimal")
	}
	component, valid := parseDecimalComponents(input)
	if !valid {
		return nil, fmt.Errorf("arrow/decimal: The string '" + input + "' is not a valid decimal number")
	}
	// Count number of significant digits (without leading zeros)
	first_non_zero := 0
	significant_digits := len(component.fractional_digits)
	whole_digits := len(component.whole_digits)
	for first_non_zero < whole_digits {
		if component.whole_digits[first_non_zero] != '0' {
			break
		}
		first_non_zero++
	}
	if first_non_zero != whole_digits {
		significant_digits += whole_digits - first_non_zero
	}

	parsed_precision := significant_digits

	parsed_scale := 0
	if component.has_exponent {
		parsed_scale = parsed_precision-1-component.exponent
	} else {
		parsed_scale = len(component.fractional_digits)
	}
	out := StringToInteger(component.whole_digits, component.fractional_digits)
	if parsed_scale < 0 {
		out.Mul(GetScaleMultiplier(-parsed_scale))
	}
	if component.sign == '-' {
		out.Negate()
	}

	if (parsed_scale < 0) {
		parsed_precision -= parsed_scale;
		parsed_scale = 0;
	}

	if (precision != nil) {
		*precision = parsed_precision;
	}
	if (scale != nil) {
		*scale = parsed_scale;
	}

	return out, nil
}

func (n Num) ToString(scale int) string {
	// TODO(fredgan): convert Decimal128 to decimal string with scale
	return ""
}


func AppendLittleEndianToArray(array []uint64, result *string) {
	arrayLen := len(array)
	most_significant_non_zero := 0
	for i := arrayLen; i > 0; i-- {
		if array[i-1] != 0 {
			most_significant_non_zero = i
			break
		}
	}
	if most_significant_non_zero == 0 {
		*result += "0"
		return
	}

	most_significant_elem_idx := most_significant_non_zero-1

    bak := make([]uint64, arrayLen)
    copy(bak, array)
	var k1e9 = uint64(1000000000)
	var kNumBits = arrayLen * 64
	// Segments will contain the array split into groups that map to decimal digits,
	// in little endian order. Each segment will hold at most 9 decimal digits.
	// For example, if the input represents 9876543210123456789, then segments will be
	// [123456789, 876543210, 9].
	// The max number of segments needed = ceil(kNumBits * log(2) / log(1e9))
	// = ceil(kNumBits / 29.897352854) <= ceil(kNumBits / 29).
	segments := make([]uint32, (kNumBits + 28) / 29)
	num_segments := 0

	most_significant_elem := most_significant_elem_idx
	for {
		// Compute remainder = copy % 1e9 and copy = copy / 1e9.
		remainder := uint32(0)
		elem := most_significant_elem
		for {
			// Compute dividend = (remainder << 32) | *elem  (a virtual 96-bit integer)
			// *elem = dividend / 1e9
			// remainder = dividend % 1e9.
			hi := bak[elem] >> 32
			lo := bak[elem] & ((uint64(1) << 32) - 1)
			dividend_hi := (uint64(remainder) << 32) | hi
			quotient_hi := dividend_hi / k1e9
			remainder = uint32(dividend_hi % k1e9)
			dividend_lo := (uint64(remainder) << 32) | lo
			quotient_lo := dividend_lo / k1e9
			remainder = uint32(dividend_lo % k1e9)
			bak[elem] = (quotient_hi << 32) | quotient_lo
			if elem == 0 {
				break
			}
			elem--
		}

		segments[num_segments] = remainder
		num_segments++
		if bak[most_significant_elem] == 0 {
			if most_significant_elem == 0 {
				break
			}
			most_significant_elem--
		}
	}

	for i := num_segments - 1; i >= 0; i-- {
		if segments[i] > 0 {
			*result += strconv.FormatUint(uint64(segments[i]),10)
		} else {
			*result += "000000000"
		}
	}
}

func (n *Num) Negate() Num {
	n.lo = ^n.lo + 1
	n.hi = ^n.hi
	if (n.lo == 0) {
		n.hi = n.hi + 1
	}
	return *n
}

func (n *Num) Add(m int64) Num {
	sum := n.LowBits() + uint64(m)
	if sum < n.LowBits() {
		n.hi = int64(n.HighBits() + 1);
	}
	n.lo = sum;
	return *n
}

func (n *Num) Mul(m Num) Num {
	L0 := uint64(n.HighBits()) >> 32
	L1 := uint64(n.HighBits()) & kIntMask
	L2 := n.LowBits() >> 32
	L3 := n.LowBits() & kIntMask

	R0 := uint64(m.HighBits()) >> 32
	R1 := uint64(m.HighBits()) & kIntMask
	R2 := m.LowBits() >> 32
	R3 := m.LowBits() & kIntMask

	product := L3 * R3
	n.lo = product & kIntMask

	sum := product >> 32

	product = L2 * R3
	sum += product
	if sum < product {
		n.hi = kCarryBit
	} else {
		n.hi = 0
	}

	product = L3 * R2
	sum += product

	n.lo += sum << 32

	if (sum < product) {
		n.hi += kCarryBit
	}

	n.hi += int64(sum >> 32)
	n.hi += int64(L1 * R3 + L2 * R2 + L3 * R1)
	n.hi += int64((L0 * R3 + L1 * R2 + L2 * R1 + L3 * R0) << 32)
	return *n
}

func (n Num) ToIntegerString() string {
	result := ""
	sign := n.Sign()
	if sign == -1 {
		result += "-"
		abs := n
		abs.Negate()

		AppendLittleEndianToArray([]uint64{abs.LowBits(), uint64(abs.HighBits())}, &result)
	} else {
		AppendLittleEndianToArray([]uint64{n.LowBits(), uint64(n.HighBits())}, &result)
	}

	return result
}

// LowBits returns the low bits of the two's complement representation of the number.
func (n Num) LowBits() uint64 { return n.lo }

// HighBits returns the high bits of the two's complement representation of the number.
func (n Num) HighBits() int64 { return n.hi }

// Sign returns:
//
//	-1 if n <  0
//	 0 if n == 0
//	+1 if n >  0
//
func (n Num) Sign() int {
	if n == (Num{}) {
		return 0
	}

	if n.HighBits() < 0 {
		return -1
	}
	return 1
}

type DecimalComponents struct {
	whole_digits string
	fractional_digits string
	exponent int
	sign uint8
	has_exponent bool
}

func IsSign(c uint8) bool { return c == '-' || c == '+' }

func IsDot(c uint8) bool { return c == '.' }

func IsDigit(c uint8) bool { return c >= '0' && c <= '9' }

func StartsExponent(c uint8) bool { return c == 'e' || c == 'E' }

func parseDigitsRun(input string, begin int, out *string) int {
	pos := begin
	for pos < len(input){
		if !IsDigit(input[pos]) {
			break
		}
		pos++
	}
	*out = input[begin:pos]
	return pos
}

func parseDecimalComponents(input string) (*DecimalComponents, bool) {
	length := len(input)
	if length == 0 {
		return nil, false
	}

	out := &DecimalComponents{}
	pos := 0

	if IsSign(input[pos]) {
		out.sign = input[pos]
		pos++
	}

	// First run of digits
	pos = parseDigitsRun(input, pos, &out.whole_digits)
	if (pos == length) {
		return out, len(out.whole_digits) != 0
	}

	// Optional dot (if given in fractional form)
	has_dot := IsDot(input[pos])
	if has_dot {
		// Second run of digits
		pos++
		pos = parseDigitsRun(input, pos, &out.fractional_digits)
	}

	if len(out.whole_digits) == 0 && len(out.fractional_digits) == 0 {
		// Need at least some digits (whole or fractional)
		return nil, false
	}

	if pos == length {
		return out, true
	}

	// Optional exponent
	if (StartsExponent(input[pos])) {
		pos++
		if (pos != length && input[pos] == '+') {
			pos++
		}
		out.has_exponent = true
		exp, err := strconv.Atoi(input[pos:])
		if err != nil {
			return nil, false
		}
		out.exponent = exp
		pos = length
	}
	return out, pos == length
}