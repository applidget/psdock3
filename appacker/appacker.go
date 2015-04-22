package appacker

// an appacker (app packer) is responsible for packaging an app into the rootfs
// it returns env required to run the app
type appacker interface {
	PackApp(rootfs, app string) error
}

type 12FactorPacker struct {

  //blablabla
}
