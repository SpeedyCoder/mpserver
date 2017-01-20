package mpserver

import(
	"golang.org/x/exp/old/netchan"
	"log"
)

func ImportChans(network, remoteaddr string, chanNames []string, 
				 dirs []netchan.Dir) [](chan Value) {
	if (len(chanNames) != len(dirs)) {
		log.Fatal("Number of channels is not equal to the number of directions.")
	}
	imp, err := netchan.Import(network, remoteaddr)
	if err != nil { log.Fatal(err) }
	res := make([](chan Value), len(dirs))

	for i := 0; i < len(dirs); i++ {
		ch := make(chan Value)
		err = imp.Import(chanNames[i], ch, dirs[i], 0)
		if err != nil { log.Fatal(err) }
		res[i] = ch

	}
	
	return res
}

func ExportChans(network, localaddr string, chanNames []string, 
				 dirs []netchan.Dir) [](chan Value) {
	if (len(chanNames) != len(dirs)) {
		log.Fatal("Number of channels is not equal to the number of directions.")
	}

	exp := netchan.NewExporter()
	res := make([](chan Value), len(dirs))

	for i := 0; i < len(dirs); i++ {
		ch := make(chan Value)
		err := exp.Export(chanNames[i], ch, dirs[i])
		if err != nil { log.Fatal(err) }
		res[i] = ch
	}

	go exp.ListenAndServe(network, localaddr)
	
	return res
}



