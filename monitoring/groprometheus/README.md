# Prometheus On-Demand Metrics

This package provides Prometheus metrics for the GroceryBot Discord bot, including on-demand metrics that are resolved by querying the database when Prometheus scrapes the metrics endpoint.

## Features

### Static Metrics

- `grocer_bot_discord_servers`: Number of Discord servers the bot is in
- `grocer_bot_command_invocations_total`: Total number of bot command invocations (labeled by command name)

### On-Demand Metrics (Database Queries)

- `grocer_bot_monthly_active_users`: Unique users who interacted in the last 30 days
- `grocer_bot_weekly_active_users`: Unique users who interacted in the last 7 days
- `grocer_bot_daily_active_users`: Unique users who interacted in the last 24 hours
- `grocer_bot_total_grocery_entries`: Total number of grocery entries
- `grocer_bot_active_guilds`: Guilds with activity in the last 30 days
- `grocer_bot_total_grocery_lists`: Total number of grocery lists
- `grocer_bot_avg_items_per_list`: Average items per grocery list

## Usage

### Basic Setup

1. Initialize metrics in your main function:

```go
// Set up database connection
db = dbUtils.Setup(dsn, logger.Named("db"), GroBotVersion)

// Set database for on-demand metrics
groprometheus.SetDB(db)

// Initialize Prometheus metrics
groprometheus.InitMetrics()
```

2. Add the metrics endpoint to your API routes:

```go
// In your API setup
e.GET("/metrics", groprometheus.PrometheusHandler())
```

3. Update static metrics as needed:

```go
// Update server count
groprometheus.UpdateDiscordServers(serverCount)

// Increment command usage
groprometheus.IncrementCommandInvocation("grolist")
```

## How It Works

- **Static Metrics**: Updated immediately when events occur (server joins, commands executed)
- **On-Demand Metrics**: Queried from the database each time Prometheus scrapes the `/metrics` endpoint
- **Custom Registry**: Uses a custom Prometheus registry to avoid conflicts with default metrics
- **Error Handling**: Database query errors don't break the metrics endpoint

## Performance Considerations

- On-demand metrics execute database queries on every scrape
- Consider caching for expensive queries if scrape frequency is high
- Use database indexes on frequently queried columns (`updated_at`, `guild_id`, etc.)
- Monitor query performance and optimize as needed

## Example Prometheus Query

```promql
# Monthly active users trend
grocer_bot_monthly_active_users

# Command usage by type
grocer_bot_command_invocations_total

# Average items per list over time
grocer_bot_avg_items_per_list
```
