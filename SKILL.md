# JujuDB CLI — Agent Skill

## Description

You have access to the `jujudb` CLI to manage a family household inventory
(freezer, fridge, pantry, etc.). You can list, create, update, and delete
items, locations, sub-locations, and categories, as well as search items
by full-text query.

## Prerequisites

The CLI must be authenticated. If any command returns "session expired" or
"not authenticated", run:

```
jujudb login --server <URL> --password <PASSWORD>
```

The session persists indefinitely on disk. You only need to login once.

## Commands Reference

### Authentication

```
jujudb login --server URL --password PASSWORD
```

Authenticates and saves the session locally. Only needed once.

### Items

List all items (with optional filters):
```
jujudb items list [--location-id ID] [--sub-location-id ID] [--category-id ID] [--limit N] [--offset N]
```

Create a new item:
```
jujudb items create --name "NAME" [--quantity N] [--location-id ID] [--sub-location-id ID] [--category-id ID] [--expiry "YYYY-MM-DD"] [--description "..."] [--notes "..."]
```

Update an existing item (only pass the fields you want to change):
```
jujudb items update ID [--name "..."] [--quantity N] [--location-id ID] [--sub-location-id ID] [--category-id ID] [--expiry "YYYY-MM-DD"] [--description "..."] [--notes "..."]
```

Delete an item:
```
jujudb items delete ID
```

### Search

Full-text search across items (names, descriptions, locations, categories, notes):
```
jujudb search "QUERY" [--location-id ID] [--sub-location-id ID] [--category-id ID] [--limit N]
```

### Locations

```
jujudb locations list
jujudb locations create --name "NAME"
jujudb locations update ID --name "NAME"
jujudb locations delete ID [--force]
```

### Sub-locations

```
jujudb sublocations list [--location-id ID]
jujudb sublocations create --name "NAME" --location-id ID
jujudb sublocations update ID [--name "..."] [--location-id ID]
jujudb sublocations delete ID [--force]
```

### Categories

```
jujudb categories list
jujudb categories create --name "NAME"
jujudb categories update ID --name "NAME"
jujudb categories delete ID [--force]
```

### Search Index Sync

Re-synchronize the MeiliSearch index with the database:
```
jujudb sync
```

## Typical Workflows

### See what's in the freezer

First, find the freezer's location ID:
```
jujudb locations list
```
Then list items for that location:
```
jujudb items list --location-id 1
```

### Search for a product

```
jujudb search "poulet"
```

### Add an item to the freezer with an expiry date

```
jujudb items create --name "Filets de saumon" --quantity 3 --location-id 1 --category-id 4 --expiry "2026-08-15"
```

### Update item quantity

```
jujudb items update 5 --quantity 2
```

### Delete an item

```
jujudb items delete 5
```

### Create a sub-location inside a location

```
jujudb sublocations create --name "Tiroir du bas" --location-id 1
```

### Force-delete a location that has items

```
jujudb locations delete 3 --force
```

## Data Conventions

- **IDs** are positive integers (1, 2, 3...)
- **Dates** use the format YYYY-MM-DD
- **Location names** are in French (Congélateur, Réfrigérateur, Garde-manger, etc.)
- **Quantity** is an integer >= 0 (default: 1)
- When `--force` is not used on delete, the CLI will report dependencies that prevent deletion

## Error Handling

| Error message | Cause | Action |
|---|---|---|
| "session expired" or "not authenticated" | Session cookie invalid or missing | Run `jujudb login --server URL --password PASSWORD` |
| "error 404: ..." | The resource ID does not exist | Check the ID with a list command |
| "Conflict: ..." | Resource has dependencies (items, sub-locations) | Delete dependencies first, or use `--force` |
| "error 400: ..." | Invalid request (missing required field, bad format) | Check required flags and value formats |

## Output Format

All output is plain text, human and LLM-readable. Items are displayed as:

```
  #ID Name
     Description: ...
     Location: LocationName > SubLocationName
     Category: CategoryName
     Quantity: N
     Expiry: YYYY-MM-DD or none
     Notes: ...
```

Locations, sub-locations, and categories are displayed as:

```
  #ID Name
```
