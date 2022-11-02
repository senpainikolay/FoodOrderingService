package main

import (
	"log"
	"sync"
)

type Rating struct {
	Mutex   sync.Mutex
	Sum     float64
	Counter float64
	Score   float64
}

func GetRatingStruct() *Rating {
	return &Rating{sync.Mutex{}, 0.0, 0.0, 0.0}
}

func (r *Rating) Add(avg_rating float64) {
	r.Mutex.Lock()

	r.Sum += avg_rating
	r.Counter += 1.0
	r.Score = r.Sum / r.Counter
	temp := r.Score
	r.Mutex.Unlock()

	log.Printf("The Rating: %v", temp)

}
