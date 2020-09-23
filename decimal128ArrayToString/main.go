package main

import (
	"errors"
	"log"
	"strconv"
)
//one 8 bytes and another 8 bytes consist to new 8 bytes, which is string.
func AppendLittleEndianarrayToString(array []uint64 ) (result string, err error) {
	//var result string
	if len(array) == 0 {
		log.Fatal("arrayay is nil")
		return result, errors.New("arrayay is nil")
	}

	//var n uint64 = uint64(len(array))
	var mostSignificantNonZero uint64
	var mostSignificantNonZeroIndex uint64
	for i := len(array) - 1; i >= 0; i-- {
		if array[i] != 0 {
			mostSignificantNonZero = array[i]
			mostSignificantNonZeroIndex = uint64(i)
			break
		}
	}

	if &mostSignificantNonZero == &array[0] {
		log.Fatal("not found")
		result = result + "0"
		return result, errors.New("the index of mostSignificantNonZero is zero.")
	}

	//3
	var mostSignificantElemIdx uint64 = mostSignificantNonZeroIndex
	var copy  = array[:mostSignificantElemIdx + 1]
	const k1e9 uint32 = 1000000000
	//var kNumBits uint64 = n * 64

	// Segments will contain the array split into groups that map to decimal digits,
	// in little endian order. Each segment will hold at most 9 decimal digits.
	// For example, if the input represents 9876543210123456789, then segments will be
	// [123456789, 876543210, 9].
	// The max number of segments needed = ceil(kNumBits * log(2) / log(1e9))
	// = ceil(kNumBits / 29.897352854) <= ceil(kNumBits / 29).
	//var segmentsLen = (kNumBits + 28) / 29
	//var segments [segmentsLen]uint32
	var segments []uint32
	var numSegments uint64 = 0
	//var mostSignificantElem *uint64 = &copy[mostSignificantElemIdx]

	var mostSignificantElem []uint64
	for i := 0; i < len(copy); i++ {
		mostSignificantElem = append(mostSignificantElem, copy[i] )
	}

	for {
		var remainder uint32 = 0

		var elem = mostSignificantElem
		var elem_len = len(elem)

		j := elem_len - 1
		var mostSignificantElemTemp []uint64
		for {
			var hi uint32 = uint32(elem[j] >> 32)
			var leastSignficantBitMask uint64 = (uint64(1) << 32) - 1
			var lo uint32 = uint32(elem[j]  & leastSignficantBitMask)
			var dividendHi uint64 = (uint64(remainder) << 32) | uint64(hi)
			var quotientHi uint64 = dividendHi / uint64(k1e9)
			remainder = uint32(dividendHi % uint64(k1e9) )
			var dividendLo uint64 = (uint64(remainder) << 32) | uint64(lo)
			var quotientLo uint64 = dividendLo / uint64(k1e9)
			remainder = uint32(dividendLo % uint64(k1e9) )
			elemTemp := (quotientHi << 32) | quotientLo
			mostSignificantElemTemp = append(mostSignificantElemTemp, elemTemp)
			if j == 0 {
				break
			}
			j--
		}
		//segments[numSegments] = remainder
		segments = append(segments, remainder)
	    numSegments++

		var elemTempLen = len(mostSignificantElemTemp)
		for i := 0; i < (elemTempLen/2); i++ {
			temp := mostSignificantElemTemp[elemTempLen - 1 - i]
			mostSignificantElemTemp[elemTempLen - 1 - i] = mostSignificantElemTemp[i]
			mostSignificantElemTemp[i] = temp
		}

		var elemTempIdx = elemTempLen - 1
	    if mostSignificantElemTemp[elemTempIdx] == 0 {
	    	if elemTempIdx == 0 {
				break
			}
			if elemTempIdx > 0 {
				elemTempIdx--
			}
		}
		mostSignificantElem = mostSignificantElemTemp[:elemTempIdx + 1]
	}

	var segment []uint32
	//add one moret time at num_segments,so delete it
	for i := 0; i < int(numSegments); i++ {
		segment = append(segment, segments[i])
	}
	log.Printf("segment=%v", segment)
	log.Println("numSegments=",numSegments)

	//var old_size uint64 = uint64(len([]byte(result)))
	//var new_size uint64 = old_size + numSegments * 9
	const newSize string = "000000000"
	var segmentLen = len(segment)
    var segmentStr string
	i := segmentLen - 1
	for ; i >= 0; i-- {
		if segment[i] > 0 {
			segmentStr += strconv.FormatUint(uint64(segment[i]),10)
		} else if segment[i] == 0 {
			segmentStr += newSize
		} else {
			return result,  errors.New("segment[i] less than zero.")
		}

	}

    i = i + 1
	for {
		if i == 0 {
			break
		}
		if i > 0 {
			j := i
			for ; j >= 0; j-- {
				segmentStr += strconv.FormatUint(uint64(segments[i]),10)
			}
		}
	}

    result = segmentStr


	return result, nil

}
//2^64 = 18446744073709551616 = 1.8446744073709552e+19
//2^128= 340282366920938463389587631136930004996 = 3.402823669209385e+38
func main() {
	//1、input the first of lower 8 bytes which is x, and the second of 8 bytes which is y.
	// example, z = x + y*2^64
	//var array = []uint64{1, 0} // 1 + 0*2^64 = 1
	//var array = []uint64{0, 1} //0 + 1*2^64 = 18446744073709551616
	//var array = []uint64{1, 1} // 1 + 1*2^64 = 18446744073709551617
	//var array = []uint64{687399551400673280, 5421010862427522170}
	//var array = []uint64{1864712049423024128, 542}
	//2. input three groups of 8 bytes
	// example, g = x + y*2^64 + z*2^128
	//var array = []uint64{1, 0, 0}
	//0 + 0*2^64 + 1*2^128 = 340282366920938463"4"63374607431768211456
	//var array = []uint64{0, 0, 1}
	//0 + 1*2^64 + 1*2^128 = 340282366920938463481821351505477763072
	//var array = []uint64{0, 1, 1}
	//1 + 1*2^64 + 1*2^128 = 340282366920938463481821351505477763073
	//var array = []uint64{1, 1, 1}
	//7766279631452241920 + 5*2^64 + 1*2^128 = 340282366920938463"5"63374607431768211456
	//var array = []uint64{7766279631452241920, 5, 1}
	// 687399551400673280 + 5421010862427522170*2^64 +1*2*128=  "4"40282366920938463463374607431768211456
	//var array = []uint64{687399551400673280, 5421010862427522170, 1}
	//2. input four groups of 8 bytes
	// example, h = x + y*2^64 + z*2^128 + g*2^192
	//0 + 0*2^64 + 0*2^128 + 1*2^192= 6277101735386680763835789423207666416102355444464034512896
	//var array = []uint64{0, 0, 0, 1}
	//0 + 0*2^64 + 1*2^128 + 1*2^192= 6277101735386680764"1"76071790128604879565730051895802724352
	//var array = []uint64{0, 0, 1, 1}
	//687399551400673280 + 5421010862427522170*2^64 + 1*2^128 + 1*2^192 = 6277101735386680764"2"7671790128604879565730051895802724352
	var array = []uint64{687399551400673280, 5421010862427522170, 1, 1}



	var result string
	result, err := AppendLittleEndianarrayToString(array)
	if err != nil {
		log.Fatal("AppendLittleEndianarrayToString is error.")
	}
	log.Println("result=",result)

}




