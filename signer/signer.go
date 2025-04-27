package main

import (
	"slices"
	"strconv"
	"strings"
	"sync"
)

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

func main() {}
