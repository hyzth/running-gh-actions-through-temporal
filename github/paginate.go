package github

import (
	"iter"

	"github.com/google/go-github/v68/github"
)

type (
	// pageFetcher represents a paginated API call
	pageFetcher[T any] interface {
		// ListOptions{Page: page} must be passed to the called API to fetch the next page
		fetch(page int) ([]T, *github.Response, error)
	}

	pageFetcherFunc[T any] func(page int) ([]T, *github.Response, error)

	// iterItem represents an item fetched from a paginated API call
	iterItem[T any] struct {
		value T
		err   error
	}
)

// newPaginatedIter creates an iterator for fetching paginated data from GitHub
func newPaginatedIter[T any](fetcher pageFetcher[T]) iter.Seq[iterItem[T]] {
	return func(yield func(iterItem[T]) bool) {
		page := 1

		for {
			items, resp, err := fetcher.fetch(page)
			if err != nil {
				yield(iterItem[T]{err: err})
				return
			}

			for _, item := range items {
				if !yield(iterItem[T]{value: item}) {
					return
				}
			}

			if resp.NextPage == 0 {
				break
			}

			page = resp.NextPage
		}
	}
}

func newPageFetcher[T any](f func(page int) ([]T, *github.Response, error)) pageFetcher[T] {
	return pageFetcherFunc[T](f)
}

func (f pageFetcherFunc[T]) fetch(page int) ([]T, *github.Response, error) {
	return f(page)
}
