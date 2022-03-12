package streaming

type Scraper interface {
	ExecuteProcess() error
}

// type Netflix struct {
// }

// type Hbo struct {
// }

// func (h Hbo) ExecuteProcess() {

// }
// func (n Netflix) ExecuteProcess() {

// }

// func main() {
// 	var scraper Scraper = Netflix{}
// 	scraper.ExecuteProcess()

// 	scraper = Hbo{}
// 	scraper.ExecuteProcess()

// }
