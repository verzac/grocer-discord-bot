---
sidebar_position: 10
---

# Legacy Message Commands (`!gro`)

This page documents the original message-command format for GroceryBot.

These commands are still supported, but slash commands are the recommended primary interface.

When using legacy message commands, mention the bot first (for example: `@GroceryBot !gro Chicken thighs`).

## Core commands

```
!gro <item name>
!grolist
!grohere
!groremove <item name>
!groremove <item index>
!groedit <item index> <new name>
!grobulk
!groclear
!grohelp
!groreset
```

## Multiple-list format (legacy)

Legacy message commands support list scoping with the `:<list-label>` suffix:

```
!gro:amazon PS5
!grolist:amazon
!groremove:amazon PS5
!grohere:amazon
!groclear:amazon
```

For list management in the legacy format:

```
!grolist new amazon My Amazon Shopping List
!grolist:amazon edit-label ebay
!grolist:amazon edit-name Kyle's Amazon Shopping List
!grolist:amazon delete
```

## `/grobulk` vs `!grobulk`

- `/grobulk` opens a modal form.
- `!grobulk` accepts free-text message input with one item per line.

Example for `!grobulk`:

```
!grobulk
Eggs
Soap 1pc
Liquid soap 500ml
```
