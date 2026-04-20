package providers

type IProvider interface {
	Validate() error
}
