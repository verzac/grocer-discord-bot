package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"github.com/verzac/grocer-discord-bot/utils"
	"go.uber.org/zap"
)

const (
	msgCannotSaveNewGroceryList = "Whoops, can't seem to save your new grocery list. Please try again later!"
	msgCmdNotFound              = ":thinking: Hmm... Not sure what you're looking for. Here are my available commands:\n`!grolist`\n`!grolist new <new list's label> <new list's fancy name - optional>`\n`!grolist:<label> delete`\n`!grolist:<label> edit-name <new fancy name>`\n`!grolist:<label> edit-label <new label>`"
	msgPrefixDefault            = "Here's your grocery list:"
)

func (m *MessageHandlerContext) OnList() error {
	if m.commandContext.ArgStr == "" {
		return m.displayList()
	}
	if strings.HasPrefix(m.commandContext.ArgStr, "new ") {
		return m.newList()
	}
	if strings.HasPrefix(m.commandContext.ArgStr, "delete") {
		return m.deleteList()
	}
	if strings.HasPrefix(m.commandContext.ArgStr, "edit-name ") {
		return m.editList()
	}
	if strings.HasPrefix(m.commandContext.ArgStr, "edit-label ") {
		return m.relabelList()
	}
	if m.commandContext.ArgStr == "all" {
		return m.displayListAll()
	}
	return m.reply(msgCmdNotFound)
}

func (m *MessageHandlerContext) displayListAll() error {
	msgPrefix := msgPrefixDefault
	groceries, err := m.groceryEntryRepo.FindByQuery(
		&models.GroceryEntry{
			GuildID: m.commandContext.GuildID,
		},
	)
	if err != nil {
		return m.onError(err)
	}
	groceryLists, err := m.groceryListRepo.FindByQuery(&models.GroceryList{GuildID: m.commandContext.GuildID})
	if err != nil {
		return m.onError(err)
	}
	if len(groceries) == 0 {
		// textBody will say something along the line of "You have no grocery lists."
		msgPrefix = ""
	}
	textBody := m.getDisplayListText(groceryLists, groceries)
	textComponents := make([]string, 0, 3)
	for _, textComponent := range []string{msgPrefix, textBody} {
		// filter empty strings
		if textComponent != "" {
			textComponents = append(textComponents, strings.TrimRight(textComponent, "\n"))
		}
	}
	return m.reply(strings.Join(textComponents, "\n"))
}

func (m *MessageHandlerContext) displayList() error {
	msgPrefix := msgPrefixDefault
	groceryList, err := m.GetGroceryListFromContext()
	if err != nil {
		return m.onGetGroceryListError(err)
	}
	groceries, err := m.groceryEntryRepo.FindByQueryWithConfig(
		&models.GroceryEntry{
			GuildID:       m.commandContext.GuildID,
			GroceryListID: groceryList.GetID(),
		},
		repositories.GroceryEntryQueryOpts{
			IsStrongNilForGroceryListID: true,
		},
	)
	if err != nil {
		return m.onError(err)
	}
	// groceryLists, err := m.groceryListRepo.FindByQuery(&models.GroceryList{GuildID: m.commandContext.GuildID})
	// if err != nil {
	// 	return m.onError(err)
	// }
	groceryLists := make([]models.GroceryList, 0, 1)
	if groceryList != nil {
		groceryLists = append(groceryLists, *groceryList)
	}
	textBody := m.getDisplayListText(groceryLists, groceries)
	if len(groceries) == 0 {
		// textBody will say something along the line of "You have no grocery lists."
		msgPrefix = ""
	}
	msgSuffix := ""
	groceryListCount, err := m.groceryListRepo.Count(&models.GroceryList{
		GuildID: m.commandContext.GuildID,
	})
	if err != nil {
		// not fatal - simply a helper to improve adoption rate
		m.LogError(err)
	} else if groceryListCount > 0 {
		msgSuffix = fmt.Sprintf("\nand %d other grocery lists (use `!grolist all`).", groceryListCount)
	}
	textComponents := make([]string, 0, 3)
	for _, textComponent := range []string{msgPrefix, textBody, msgSuffix} {
		// filter empty strings
		if textComponent != "" {
			textComponents = append(textComponents, strings.TrimRight(textComponent, "\n"))
		}
	}
	return m.reply(strings.Join(textComponents, "\n"))
}

func (m *MessageHandlerContext) getDisplayListText(groceryLists []models.GroceryList, groceries []models.GroceryEntry) string {
	// group by their grocerylist
	if len(groceryLists) == 0 && len(groceries) == 0 {
		return getNoGroceryText("")
	}
	noListGroceries, groupedGroceries, listlessGroceries := utils.GroupByGroceryLists(groceryLists, groceries)
	for _, g := range listlessGroceries {
		m.LogError(
			errors.New("unknown grocery list ID in grocery entry"),
			zap.Uint("GroceryListID", *g.GroceryListID),
		)
	}
	noListGroceriesTxt := getGroceryListText(noListGroceries, nil)
	labeledGroceriesTxt := ""
	for _, groceryList := range groceryLists {
		g := groupedGroceries[groceryList.ID]
		label := groceryList.ListLabel
		fancyName := groceryList.FancyName
		groceryListText := ""
		if fancyName != nil {
			groceryListText = fmt.Sprintf("**%s** (%s)", *fancyName, label)
		} else {
			groceryListText = fmt.Sprintf("**%s**", label)
		}
		labeledGroceriesTxt += fmt.Sprintf(":shopping_cart: %s\n%s\n", groceryListText, getGroceryListText(g, &groceryList))
	}
	return strings.Join([]string{noListGroceriesTxt, labeledGroceriesTxt}, "\n")
}

func (m *MessageHandlerContext) newList() error {
	if m.commandContext.GrocerySublist != "" {
		return m.reply(msgCmdNotFound)
	}
	splitArgs := strings.SplitN(m.commandContext.ArgStr, " ", 3)
	if len(splitArgs) < 2 {
		return m.reply("Sorry, I need to know what you'd like to label your new grocery list as. For example, you can type `!grolist new amazon` to make a grocery list with the label `amazon`.")
	}
	label := splitArgs[1]
	var fancyName string
	if len(splitArgs) >= 3 && splitArgs[2] != "" {
		fancyName = splitArgs[2]
	}
	existingCountInGuild, err := m.groceryListRepo.Count(&models.GroceryList{GuildID: m.commandContext.GuildID})
	if err != nil {
		return m.onError(err)
	}
	maxGroceryListPerServer := m.getMaxGroceryListPerServer()
	if existingCountInGuild+1 >= int64(maxGroceryListPerServer) {
		m.logger.Info(
			"maxGroceryListPerServer limit hit.",
			zap.String("GuildID", m.commandContext.GuildID),
			zap.Int("maxGroceryListPerServer", maxGroceryListPerServer),
		)
		return m.reply(fmt.Sprintf(
			":shopping_bags: Whoops, you already have the maximum of %d grocery lists for this server. Please delete one through `!grolist delete <list label>` to make room for new ones. Alternatively, you can use `!grolist edit-label` and `!grolist edit-name` to edit your grocery list (see `!grohelp` for more details).\n\nPS: You can get higher limits by supporting me on Patreon through `!grohelp` or `/grohelp`!",
			existingCountInGuild+1,
		))
	}
	newGroceryList, err := m.groceryListRepo.CreateGroceryList(m.commandContext.GuildID, label, fancyName)
	if err != nil {
		switch err {
		case repositories.ErrGroceryListDuplicate:
			return m.reply(fmt.Sprintf("Sorry, a grocery list with the label **%s** already exists for your server. Please select another label :)", label))
		default:
			m.LogError(err)
			return m.reply(msgCannotSaveNewGroceryList)
		}
	}
	if err := m.reply(fmt.Sprintf("Yay! Your new grocery list **%s** has been successfully created. Use it in a command like so to add entries to your grocery list: `!gro:%s Chicken`", newGroceryList.GetName(), newGroceryList.ListLabel)); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}

func fmtErrGroceryListNotFound(label string) string {
	return fmt.Sprintf("Whoops, I cannot seem to find a grocery list with the label **%s**... Could you please try again?", label)
}

func (m *MessageHandlerContext) deleteList() error {
	label := m.commandContext.GrocerySublist
	if label == "" {
		return m.reply(msgCmdNotFound)
	}
	groceryList, err := m.groceryListRepo.GetByQuery(&models.GroceryList{ListLabel: label})
	if err != nil {
		return m.onError(err)
	}
	if groceryList == nil {
		return m.reply(fmtErrGroceryListNotFound(label))
	}
	count, rErr := m.groceryEntryRepo.GetCount(&models.GroceryEntry{GuildID: m.commandContext.GuildID, GroceryListID: &groceryList.ID})
	if rErr != nil {
		return m.onError(rErr)
	}
	if count > 0 {
		return m.reply(fmt.Sprintf("Oops, you still have %d groceries in **%s**. Pro-tip: use `!groclear:%s` to clear your groceries!", count, groceryList.GetName(), groceryList.ListLabel))
	}
	if err := m.onListRemoveGrohereRecord(groceryList); err != nil {
		return m.onError(err)
	}
	if err := m.groceryListRepo.Delete(groceryList); err != nil {
		switch err {
		case repositories.ErrGroceryListNotFound:
			return m.reply(fmtErrGroceryListNotFound(label))
		default:
			return m.onError(err)
		}
	}
	if err := m.reply(fmt.Sprintf("Successfully deleted **%s**! Feel free to make new ones with `!grolist new`!", label)); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohere()
}

func (m *MessageHandlerContext) editList() error {
	label := m.commandContext.GrocerySublist
	if label == "" {
		return m.reply(msgCmdNotFound)
	}
	splitArgs := strings.SplitN(m.commandContext.ArgStr, " ", 2)
	if len(splitArgs) < 2 {
		return m.reply("Sorry, I need to know what you'd like to rename your grocery as. For example: `!grolist:amazon edit-name My Amazon Shopping List` to change a grocery list with the label `amazon` to have the name \"My Amazon Shopping List\". Changing the labels themselves are done through edit-label like so: `!grolist edit-label amazon ebay`.")
	}
	newFancyName := splitArgs[1]
	groceryList, err := m.groceryListRepo.GetByQuery(&models.GroceryList{ListLabel: label})
	if err != nil {
		return m.onError(err)
	}
	if groceryList == nil {
		return m.reply(fmt.Sprintf("Whoops, can't seem to find a grocery list with the label **%s**. You can make the grocery list by typing `!grolist new %s %s`.", label, label, newFancyName))
	}
	groceryList.FancyName = &newFancyName
	if err := m.groceryListRepo.Save(groceryList); err != nil {
		return m.onError(err)
	}
	if err := m.reply(fmt.Sprintf("Successfully edited grocery list with the label **%s** to have the following name: %s.", label, *groceryList.FancyName)); err != nil {
		return m.onError(err)
	}
	return m.onEditUpdateGrohereWithGroceryList()
}

func (m *MessageHandlerContext) relabelList() error {
	label := m.commandContext.GrocerySublist
	if label == "" {
		return m.reply(msgCmdNotFound)
	}
	splitArgs := strings.SplitN(m.commandContext.ArgStr, " ", 2)
	if len(splitArgs) < 2 {
		return m.reply("Sorry, I need to know what you'd like to relabel your grocery list as. For example: `!grolist:amazon edit-label ebay` to change your `ebay` list's label to `amazon instead` (all items will be moved to the new list).")
	}
	newLabel := splitArgs[1]
	groceryList, err := m.groceryListRepo.GetByQuery(&models.GroceryList{ListLabel: label})
	if err != nil {
		return m.onError(err)
	}
	if groceryList == nil {
		return m.reply(fmt.Sprintf("Whoops, can't seem to find a grocery list with the label **%s**. You can make the grocery list by typing `!grolist new %s My Shopping List`.", label, newLabel))
	}
	groceryList.ListLabel = newLabel
	if err := m.groceryListRepo.Save(groceryList); err != nil {
		return m.onError(err)
	}
	if err := m.reply(fmt.Sprintf(
		"Successfully edited grocery list with the label **%s** to have the following label: **%s**. Please ensure that you use commands with the new label! For example: `!gro:%s Chicken strips`",
		label,
		groceryList.ListLabel,
		groceryList.ListLabel,
	)); err != nil {
		return m.onError(err)
	}
	// since it's re-labeled, have commandContext be updated with the new label
	m.commandContext.GrocerySublist = newLabel
	return m.onEditUpdateGrohereWithGroceryList()
}

func getGroceryListText(groceries []models.GroceryEntry, groceryList *models.GroceryList) string {
	if groceryList != nil && len(groceries) == 0 {
		label := ""
		if groceryList != nil {
			label = ":" + groceryList.ListLabel
		}
		return getNoGroceryText(label)
	}
	msg := ""
	for i, grocery := range groceries {
		msg += fmt.Sprintf("%d: %s\n", i+1, grocery.ItemDesc)
	}
	return msg
}

func getNoGroceryText(label string) string {
	return fmt.Sprintf("You have no groceries - add one through `!gro` (e.g. `!gro%s Toilet paper`)!\n", label)
}
