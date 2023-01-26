package main

import (
	"github.com/charmbracelet/bubbles/list"
)

func takeFirst[S any](seq []S, n int) []S {
	if len(seq) < n {
		return seq
	}
	return seq[0:n]
}

func takeFirstString(seq string, n int) string {
	if len(seq) < n {
		return seq
	}
	return seq[0:n]
}

func insertItemIntoDescendingList(sortedArray []list.Item, event item) []list.Item {
	start := 0
	end := len(sortedArray) - 1
	var mid int
	position := start

	if end < 0 {
		return []list.Item{event}
	} else if event.CreatedAt.Before(sortedArray[end].(item).CreatedAt) {
		return append(sortedArray, event)
	} else if event.CreatedAt.After(sortedArray[start].(item).CreatedAt) {
		newArr := make([]list.Item, len(sortedArray)+1)
		newArr[0] = event
		copy(newArr[1:], sortedArray)
		return newArr
	} else if event.CreatedAt.Equal(sortedArray[start].(item).CreatedAt) {
		position = start
	} else {
		for {
			if end <= start+1 {
				position = end
				break
			}
			mid = int(start + (end-start)/2)
			if sortedArray[mid].(item).CreatedAt.After(event.CreatedAt) {
				start = mid
			} else if sortedArray[mid].(item).CreatedAt.Before(event.CreatedAt) {
				end = mid
			} else {
				position = mid
				break
			}
		}
	}

	if sortedArray[position].(item).ID != event.ID {
		newArr := make([]list.Item, len(sortedArray)+1)
		copy(newArr[:position], sortedArray[:position])
		newArr[position] = event
		copy(newArr[position+1:], sortedArray[position:])
		return newArr
	}

	return sortedArray
}
