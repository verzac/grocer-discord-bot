package groceryutils

import "fmt"

func getNoGroceryText(label string) string {
	return fmt.Sprintf("You have no groceries - add one through `!gro` (e.g. `!gro%s Toilet paper`)!\n", label)
}
