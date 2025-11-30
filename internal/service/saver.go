package service

import (
	"github.com/Liana-wq1/my-first-go/internal/model"
	"github.com/Liana-wq1/my-first-go/internal/repository"
)

func StartSaver(ch <-chan model.Entity) {
	for e := range ch {
		repository.SaveEntity(e)
	}
}
