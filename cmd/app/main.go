import (
	"github.com/mezsh77/astragen/internal/gsheets"
	"github.com/mezsh77/astragen/pkg/models"
)

func main() {
	ctx := context.Background()
	creds, err := os.ReadFile("credentials.json")
	check(err)
	sheetService, err := gsheets.NewService(ctx, creds)
	check(err)
	err = gsheets.RunSync(ctx)
}

func check(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}
