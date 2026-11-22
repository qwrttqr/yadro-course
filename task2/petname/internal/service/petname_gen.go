package service

import petname "github.com/dustinkirkland/golang-petname"

type PetnameGenerationService struct {
}

func NewPetNameGenerationService() *PetnameGenerationService {
	return &PetnameGenerationService{}
}

func (p *PetnameGenerationService) Generate(words int64, separator string) string {
	return petname.Generate(int(words), separator)
}
