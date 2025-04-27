package main

import (
	"slices"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup

	in := make(chan any)
	for i := range jobs {
		out := make(chan any)

		wg.Add(1)
		go func(j job, in, out chan any) {
			defer wg.Done()
			defer close(out)
			jobs[i](in, out)
		}(jobs[i], in, out)
		in = out
	}

	wg.Wait()
}

func SingleHash(in, out chan any) {
	for data := range in {
		dataString := strconv.Itoa(data.(int))

		parts := make([]string, 2)

		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			parts[0] = DataSignerCrc32(dataString) // 1s
			wg.Done()
		}()

		md5Digest := DataSignerMd5(dataString) // 10ms

		go func() {
			parts[1] = DataSignerCrc32(md5Digest) // 1s
			wg.Done()
		}()

		wg.Wait()

		out <- strings.Join(parts, "~")
	}
}

func MultiHash(in, out chan any) {
	for data := range in {
		dataString := data.(string)

		parts := make([]string, 6)

		var wg sync.WaitGroup

		wg.Add(6)
		for i := range 6 {
			go func(i int) {
				defer wg.Done()
				parts[i] = DataSignerCrc32(strconv.Itoa(i) + dataString)
			}(i)
		}

		wg.Wait()

		out <- strings.Join(parts, "")
	}
}

func CombineResults(in, out chan any) {
	result := make([]string, 0)

	for data := range in {
		result = append(result, data.(string))
	}

	slices.Sort(result)

	out <- strings.Join(result, "_")
}

func main() {}
