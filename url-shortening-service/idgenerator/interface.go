package idgenerator

type IDGeneratorInterface interface {
	GenerateShortCode() (string, error)
}
