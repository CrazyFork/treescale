package tree_lib

import (
	"reflect"
	"os"
	"time"
	"math/rand"
	"io"
	"math/big"
	"math"
)


func GetEnv(name, def_value string) string {
	val := os.Getenv(name)
	if len(val) > 0 {
		return val
	}
	return def_value
}

func ArrayContains(array interface{}, v interface{}) (int, bool) {
	if reflect.TypeOf(array).Kind() != reflect.Slice && reflect.TypeOf(array).Kind() != reflect.Array {
		return -1, false
	}
	arr_val := reflect.ValueOf(array)

	if arr_val.Len() == 0 {
		return -1, false
	}

	// Testing element types
	if reflect.TypeOf(arr_val.Index(0).Interface()).Kind() != reflect.TypeOf(v).Kind() {
		return -1, false
	}

	// Trying to find value
	for i:=0; i< arr_val.Len(); i++ {
		if arr_val.Index(i).Interface() == v {
			return i, true
		}
	}

	return -1, false
}


// Function checks is 2 Arrays contains same element or not
// And returning indexes for both, with bool containing or not
func ArrayMatchElement(array1 interface{}, array2 interface{}) (int, int, bool) {
	if	(reflect.TypeOf(array1).Kind() != reflect.Slice && reflect.TypeOf(array1).Kind() != reflect.Array) ||
		(reflect.TypeOf(array2).Kind() != reflect.Slice && reflect.TypeOf(array2).Kind() != reflect.Array) {
		return -1, -1, false
	}

	v1 := reflect.ValueOf(array1)
	v2 := reflect.ValueOf(array2)

	if v1.Len() == 0 || v2.Len() == 0 {
		return -1, -1, false
	}

	if reflect.TypeOf(v1.Index(0).Interface()).Kind() != reflect.TypeOf(v2.Index(0).Interface()).Kind() {
		return -1, -1, false
	}

	for i:=0; i< v1.Len(); i++ {
		for j:=0; j< v1.Len(); j++ {
			if v1.Index(i).Interface() == v2.Index(j).Interface() {
				return i, j, true
			}
		}
	}

	return -1, -1, false
}


const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789!&$#*-_~()}{][|"
	file_letters_Bytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
)

func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandomFileName(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = file_letters_Bytes[rand.Intn(len(file_letters_Bytes))]
	}
	return string(b)
}

func CopyFile(src, dst string) (err TreeError) {
	var (
		db_f, new_db_f	*os.File
	)
	err.From = FROM_COPY_FILE
	db_f, err.Err = os.Open(src)
	if !err.IsNull() {
		return
	}
	defer db_f.Close()

	new_db_f, err.Err = os.Create(dst)
	if !err.IsNull() {
		return
	}
	defer new_db_f.Close()

	_, err.Err = io.Copy(new_db_f, db_f)
	return
}

func NextPrimeNumber (n int64) int64 {
	var (
		i 		= 		n+1
		j				int64
		mark 			bool
	)
	for {
		mark = true
		for j = 2; j<=int64(math.Sqrt(float64(n))); j++ {
			if i % j == 0 {
				mark = false
				break
			}
		}
		if mark {
			return i
		}
		i++
	}
	return i
}

// If y dividable to x without without remains
// then it  returns true and pointer to divided x parameter with final result
func IsBigDividable(x, y *big.Int) (bool, *big.Int) {
	div := big.Int{}
	div.Mod(x, y)
	if div.Int64() == 0 {
		x.Div(x, y)
		return true, x
	}
	
	return false, x
}

func LCM (a *big.Int, b *big.Int) (c *big.Int) {
	x := big.Int{}
	c = big.NewInt(1)
	c.Mul(a,b)
	x.GCD(nil, nil , a, b)
	c.Div(c,&x)
	return
}