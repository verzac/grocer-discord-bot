package ingredients

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/services/grocery"
	"github.com/verzac/grocer-discord-bot/services/registration"
	"go.uber.org/zap"
)

func (s *IngredientsServiceImpl) StorePending(ingredients []string, guildID, authorID, listLabel string) string {
	return s.memCache.set(&pendingIngredients{
		Ingredients: ingredients,
		GuildID:     guildID,
		AuthorID:    authorID,
		ListLabel:   strings.TrimSpace(listLabel),
	})
}

func (s *IngredientsServiceImpl) Cancel(cacheKey string) {
	s.memCache.delete(cacheKey)
}

func (s *IngredientsServiceImpl) PendingAuthorID(cacheKey string) (authorID string, ok bool) {
	p, ok := s.memCache.peek(cacheKey)
	if !ok || p == nil {
		return "", false
	}
	return p.AuthorID, true
}

func (s *IngredientsServiceImpl) ConfirmAndAdd(ctx context.Context, cacheKey, authorID string) (addedCount int, err error) {
	p0, ok := s.memCache.peek(cacheKey)
	if !ok {
		return 0, ErrPendingNotFound
	}
	if p0.AuthorID != authorID {
		return 0, ErrWrongAuthor
	}
	keyGuild, _, keyHasGuild := strings.Cut(cacheKey, ":")
	if !keyHasGuild || keyGuild != p0.GuildID {
		return 0, ErrPendingNotFound
	}

	groceryList, err := s.resolveGroceryList(ctx, p0.GuildID, p0.ListLabel)
	if err != nil {
		return 0, err
	}

	toInsert := make([]models.GroceryEntry, 0, len(p0.Ingredients))
	for _, item := range p0.Ingredients {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		aID := p0.AuthorID
		toInsert = append(toInsert, models.GroceryEntry{
			ItemDesc:    item,
			GuildID:     p0.GuildID,
			UpdatedByID: &aID,
		})
	}
	if len(toInsert) == 0 {
		return 0, errors.New("no ingredients to add")
	}

	registrationContext, regErr := registration.Service.GetRegistrationContext(p0.GuildID)
	if regErr != nil {
		s.logger.Error("registration lookup failed", zap.Error(regErr))
	}

	limitOk, groceryEntryLimit, err := grocery.Service.ValidateGroceryEntryLimit(ctx, registrationContext, p0.GuildID, len(toInsert))
	if err != nil {
		return 0, err
	}
	if !limitOk {
		return 0, fmt.Errorf("Whoops, you've gone over the limit allowed by the bot (max %d grocery entries per server). Please log an issue through GitHub (look at `!grohelp`) to request an increase! Thank you for being a power user! :tada:", groceryEntryLimit)
	}

	pending, ok := s.memCache.take(cacheKey)
	if !ok {
		return 0, ErrPendingNotFound
	}

	rErr := s.groceryEntryRepo.WithContext(ctx).AddToGroceryList(groceryList, toInsert, pending.GuildID)
	if rErr != nil {
		return 0, fmt.Errorf("%s", rErr.Message)
	}

	if err := grocery.Service.OnGroceryListEdit(ctx, groceryList, pending.GuildID); err != nil {
		return len(toInsert), err
	}
	return len(toInsert), nil
}

func (s *IngredientsServiceImpl) resolveGroceryList(ctx context.Context, guildID, listLabel string) (*models.GroceryList, error) {
	if listLabel == "" {
		return nil, nil
	}
	gl, err := s.groceryListRepo.WithContext(ctx).GetByQuery(&models.GroceryList{ListLabel: listLabel, GuildID: guildID})
	if err != nil {
		return nil, err
	}
	if gl == nil {
		return nil, fmt.Errorf("Whoops, I can't seem to find the grocery list labeled as *%s*.", listLabel)
	}
	return gl, nil
}
