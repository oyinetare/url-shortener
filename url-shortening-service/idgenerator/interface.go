package idgenerator

type IDGeneratorInterface interface {
	GenerateShortCode(longUrl string) (string, error)
}
