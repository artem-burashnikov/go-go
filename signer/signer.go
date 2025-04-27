package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"hash/crc32"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type job func(in, out chan any)

const (
	MaxInputDataLen = 100
)

var (
	dataSignerOverheat uint32 = 0
	DataSignerSalt            = ""
)

var OverheatLock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 0, 1); !swapped {
			fmt.Println("OverheatLock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var OverheatUnlock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 1, 0); !swapped {
			fmt.Println("OverheatUnlock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var DataSignerMd5 = func(data string) string {
	OverheatLock()
	defer OverheatUnlock()
	data += DataSignerSalt
	dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	time.Sleep(10 * time.Millisecond)
	return dataHash
}

var DataSignerCrc32 = func(data string) string {
	data += DataSignerSalt
	crcH := crc32.ChecksumIEEE([]byte(data))
	dataHash := strconv.FormatUint(uint64(crcH), 10)
	time.Sleep(time.Second)
	return dataHash
}

func makeJob(j job, in chan any) chan any {
	out := make(chan any)
	go func() {
		defer close(out)
		j(in, out)
	}()
	return out
}

func ExecutePipeline(jobs ...job) {
	var in chan any
	for _, j := range jobs {
		in = makeJob(j, in)
	}

	// This spins while the last out channel of the last job is open.
	for range in {
	}
}

func SingleHash(in, out chan any) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for rawData := range in {
		stringData := strconv.Itoa(rawData.(int))

		wg.Add(1)
		go func(stringData string) {
			defer wg.Done()
			parts := make([]string, 2)

			var inner sync.WaitGroup

			inner.Add(1)
			go func() {
				defer inner.Done()
				parts[0] = DataSignerCrc32(stringData)
			}()

			inner.Add(1)
			go func() {
				defer inner.Done()
				mu.Lock()
				md5Digest := DataSignerMd5(stringData)
				mu.Unlock()
				parts[1] = DataSignerCrc32(md5Digest)
			}()

			inner.Wait()
			out <- strings.Join(parts, "~")
		}(stringData)
	}

	wg.Wait()
}

func MultiHash(in, out chan any) {
	var wg sync.WaitGroup

	for data := range in {
		dataString := data.(string)

		wg.Add(1)
		go func(dataString string) {
			defer wg.Done()
			parts := make([]string, 6)

			var inner sync.WaitGroup

			for i := range 6 {
				inner.Add(1)
				go func(i int) {
					defer inner.Done()
					parts[i] = DataSignerCrc32(strconv.Itoa(i) + dataString)
				}(i)
			}

			inner.Wait()

			out <- strings.Join(parts, "")
		}(dataString)
	}

	wg.Wait()
}

func CombineResults(in, out chan any) {
	result := make([]string, 0, MaxInputDataLen)

	for data := range in {
		result = append(result, data.(string))
	}

	slices.Sort(result)

	out <- strings.Join(result, "_")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: signer <salt> <number1> [number2] ... [numberN]")
		os.Exit(1)
	}

	DataSignerSalt = args[0]

	input := make([]int, 0, len(args)-1)

	for _, arg := range args[1:] {
		num, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid number %q: %v\n", arg, err)
			os.Exit(1)
		}
		input = append(input, num)
	}

	hashSignJobs := []job{
		job(func(in, out chan any) {
			for _, num := range input {
				out <- num
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan any) {
			fmt.Println((<-in).(string))
		}),
	}

	ExecutePipeline(hashSignJobs...)
}
