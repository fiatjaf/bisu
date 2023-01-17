package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/nbd-wtf/go-nostr"
)

func insertItemIntoDescendingList(sortedArray []list.Item, event item) []list.Item {
	start := 0
	end := len(sortedArray) - 1
	midPoint := 0
	position := start

	if end < 0 {
		return []list.Item{event}
	} else if event.CreatedAt.Before(sortedArray[end].(item).CreatedAt) {
		position = end + 1
	} else if !event.CreatedAt.Before(sortedArray[start].(item).CreatedAt) {
		position = start + 1
	} else {
		for {
			if end <= start+1 {
				position = end
				break
			}
			midPoint = int(start + (end-start)/2)
			if sortedArray[midPoint].(item).CreatedAt.After(event.CreatedAt) {
				start = midPoint
			} else if sortedArray[midPoint].(item).CreatedAt.Before(event.CreatedAt) {
				end = midPoint
			} else {
				position = midPoint
				break
			}
		}
	}

	if sortedArray[position-1].(item).ID != event.ID {
		newArr := make([]list.Item, len(sortedArray)+1)
		copy(newArr[:position], sortedArray[:position])
		newArr[position] = event
		copy(newArr[position+1:], sortedArray[position:])
		return newArr
	}

	return sortedArray
}

func insertEventIntoDescendingList(sortedArray []*nostr.Event, event *nostr.Event) []*nostr.Event {
	start := 0
	end := len(sortedArray) - 1
	midPoint := 0
	position := start

	if end < 0 {
		position = 0
	} else if event.CreatedAt.Before(sortedArray[end].CreatedAt) {
		position = end + 1
	} else if !event.CreatedAt.Before(sortedArray[start].CreatedAt) {
		position = start
	} else {
		for {
			if end <= start+1 {
				position = end
				break
			}
			midPoint = int(start + (end-start)/2)
			if sortedArray[midPoint].CreatedAt.After(event.CreatedAt) {
				start = midPoint
			} else if sortedArray[midPoint].CreatedAt.Before(event.CreatedAt) {
				end = midPoint
			} else {
				position = midPoint
				break
			}
		}
	}

	if sortedArray[position].ID != event.ID {
		newArr := make([]*nostr.Event, len(sortedArray)+1)
		copy(newArr[:position], sortedArray[:position])
		newArr[position] = event
		copy(newArr[position+1:], sortedArray[position:])
		return newArr
	}

	return sortedArray
}
