package groprometheus

import (
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

var (
	// Custom registry
	registry = prometheus.NewRegistry()
	hasInit  = atomic.Bool{}

	// Database connection for on-demand metrics
	db *gorm.DB

	// Simple metric to count Discord servers
	discordServersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "grocer_bot_discord_servers",
			Help: "Number of Discord servers the bot is in",
		},
	)

	// Counter to track bot command invocations
	commandInvocationCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grocer_bot_command_invocations_total",
			Help: "Total number of bot command invocations",
		},
		[]string{"command_name"},
	)

	metricsCache = cache.New(5*time.Minute, 10*time.Minute)
)

// OnDemandCollector implements prometheus.Collector for metrics that need database queries
type OnDemandCollector struct {
	db *gorm.DB

	// Metrics
	monthlyActiveUsersGauge  *prometheus.Desc
	weeklyActiveUsersGauge   *prometheus.Desc
	dailyActiveUsersGauge    *prometheus.Desc
	totalGroceryEntriesGauge *prometheus.Desc
	activeGuildsGauge        *prometheus.Desc
	totalGroceryListsGauge   *prometheus.Desc
	avgItemsPerListGauge     *prometheus.Desc
}

// NewOnDemandCollector creates a new collector for on-demand metrics
func NewOnDemandCollector(database *gorm.DB) *OnDemandCollector {
	return &OnDemandCollector{
		db: database,
		monthlyActiveUsersGauge: prometheus.NewDesc(
			"grocer_bot_monthly_active_users",
			"Number of unique users who have interacted with the bot in the last 30 days",
			nil, nil,
		),
		weeklyActiveUsersGauge: prometheus.NewDesc(
			"grocer_bot_weekly_active_users",
			"Number of unique users who have interacted with the bot in the last 7 days",
			nil, nil,
		),
		dailyActiveUsersGauge: prometheus.NewDesc(
			"grocer_bot_daily_active_users",
			"Number of unique users who have interacted with the bot in the last 24 hours",
			nil, nil,
		),
		totalGroceryEntriesGauge: prometheus.NewDesc(
			"grocer_bot_total_grocery_entries",
			"Total number of grocery entries in the database",
			nil, nil,
		),
		activeGuildsGauge: prometheus.NewDesc(
			"grocer_bot_active_guilds",
			"Number of guilds with grocery entries in the last 30 days",
			nil, nil,
		),
		totalGroceryListsGauge: prometheus.NewDesc(
			"grocer_bot_total_grocery_lists",
			"Total number of grocery lists in the database",
			nil, nil,
		),
		avgItemsPerListGauge: prometheus.NewDesc(
			"grocer_bot_avg_items_per_list",
			"Average number of items per grocery list",
			nil, nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *OnDemandCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.monthlyActiveUsersGauge
	ch <- c.weeklyActiveUsersGauge
	ch <- c.dailyActiveUsersGauge
	ch <- c.totalGroceryEntriesGauge
	ch <- c.activeGuildsGauge
	ch <- c.totalGroceryListsGauge
	ch <- c.avgItemsPerListGauge
}

// Collect implements prometheus.Collector
func (c *OnDemandCollector) Collect(ch chan<- prometheus.Metric) {
	if c.db == nil {
		return
	}

	// Collect all on-demand metrics
	c.collectActiveUsers(ch)
	c.collectTotalGroceryEntries(ch)
	c.collectActiveGuilds(ch)
	c.collectGroceryListStats(ch)
}

// collectActiveUsers collects daily, weekly, and monthly active user counts
func (c *OnDemandCollector) collectActiveUsers(ch chan<- prometheus.Metric) {
	now := time.Now()

	// Daily active users (last 24 hours)
	dailyKey := "daily_active_users"
	if cached, found := metricsCache.Get(dailyKey); found {
		ch <- prometheus.MustNewConstMetric(c.dailyActiveUsersGauge, prometheus.GaugeValue, cached.(float64))
	} else {
		dailyCount := c.getActiveUserCount(now.AddDate(0, 0, -1))
		metricsCache.Set(dailyKey, float64(dailyCount), cache.DefaultExpiration)
		ch <- prometheus.MustNewConstMetric(c.dailyActiveUsersGauge, prometheus.GaugeValue, float64(dailyCount))
	}

	// Weekly active users (last 7 days)
	weeklyKey := "weekly_active_users"
	if cached, found := metricsCache.Get(weeklyKey); found {
		ch <- prometheus.MustNewConstMetric(c.weeklyActiveUsersGauge, prometheus.GaugeValue, cached.(float64))
	} else {
		weeklyCount := c.getActiveUserCount(now.AddDate(0, 0, -7))
		metricsCache.Set(weeklyKey, float64(weeklyCount), cache.DefaultExpiration)
		ch <- prometheus.MustNewConstMetric(c.weeklyActiveUsersGauge, prometheus.GaugeValue, float64(weeklyCount))
	}

	// Monthly active users (last 30 days)
	monthlyKey := "monthly_active_users"
	if cached, found := metricsCache.Get(monthlyKey); found {
		ch <- prometheus.MustNewConstMetric(c.monthlyActiveUsersGauge, prometheus.GaugeValue, cached.(float64))
	} else {
		monthlyCount := c.getActiveUserCount(now.AddDate(0, 0, -30))
		metricsCache.Set(monthlyKey, float64(monthlyCount), cache.DefaultExpiration)
		ch <- prometheus.MustNewConstMetric(c.monthlyActiveUsersGauge, prometheus.GaugeValue, float64(monthlyCount))
	}
}

// getActiveUserCount counts unique users who have interacted since the given time
func (c *OnDemandCollector) getActiveUserCount(since time.Time) int64 {
	var count int64

	// Count unique users who have created or updated grocery entries since the given time
	err := c.db.Model(&struct {
		UpdatedByID string
	}{}).
		Select("COUNT(DISTINCT updated_by_id)").
		Joins("JOIN grocery_entries ON grocery_entries.updated_by_id = updated_by_id").
		Where("grocery_entries.updated_at >= ? AND grocery_entries.updated_by_id IS NOT NULL", since).
		Count(&count).Error

	if err != nil {
		// Return 0 on error, but in production you might want to log this
		return 0
	}

	return count
}

// collectTotalGroceryEntries collects the total number of grocery entries
func (c *OnDemandCollector) collectTotalGroceryEntries(ch chan<- prometheus.Metric) {
	key := "total_grocery_entries"
	if cached, found := metricsCache.Get(key); found {
		ch <- prometheus.MustNewConstMetric(c.totalGroceryEntriesGauge, prometheus.GaugeValue, cached.(float64))
		return
	}

	var count int64
	err := c.db.Model(&struct {
		ID uint
	}{}).
		Select("COUNT(*)").
		Joins("JOIN grocery_entries ON grocery_entries.id = id").
		Count(&count).Error

	if err != nil {
		return
	}

	metricsCache.Set(key, float64(count), cache.DefaultExpiration)
	ch <- prometheus.MustNewConstMetric(c.totalGroceryEntriesGauge, prometheus.GaugeValue, float64(count))
}

// collectActiveGuilds collects the number of guilds with recent activity
func (c *OnDemandCollector) collectActiveGuilds(ch chan<- prometheus.Metric) {
	key := "active_guilds"
	if cached, found := metricsCache.Get(key); found {
		ch <- prometheus.MustNewConstMetric(c.activeGuildsGauge, prometheus.GaugeValue, cached.(float64))
		return
	}

	var count int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	err := c.db.Model(&struct {
		GuildID string
	}{}).
		Select("COUNT(DISTINCT guild_id)").
		Joins("JOIN grocery_entries ON grocery_entries.guild_id = guild_id").
		Where("grocery_entries.updated_at >= ?", thirtyDaysAgo).
		Count(&count).Error

	if err != nil {
		return
	}

	metricsCache.Set(key, float64(count), cache.DefaultExpiration)
	ch <- prometheus.MustNewConstMetric(c.activeGuildsGauge, prometheus.GaugeValue, float64(count))
}

// collectGroceryListStats collects grocery list statistics
func (c *OnDemandCollector) collectGroceryListStats(ch chan<- prometheus.Metric) {
	// Count total grocery lists
	listCountKey := "total_grocery_lists"
	var listCount int64
	if cached, found := metricsCache.Get(listCountKey); found {
		ch <- prometheus.MustNewConstMetric(c.totalGroceryListsGauge, prometheus.GaugeValue, cached.(float64))
	} else {
		err := c.db.Model(&struct {
			ID uint
		}{}).
			Select("COUNT(*)").
			Joins("JOIN grocery_lists ON grocery_lists.id = id").
			Count(&listCount).Error

		if err != nil {
			return
		}

		metricsCache.Set(listCountKey, float64(listCount), cache.DefaultExpiration)
		ch <- prometheus.MustNewConstMetric(c.totalGroceryListsGauge, prometheus.GaugeValue, float64(listCount))
	}

	// Calculate average items per list
	avgItemsKey := "avg_items_per_list"
	if cached, found := metricsCache.Get(avgItemsKey); found {
		ch <- prometheus.MustNewConstMetric(c.avgItemsPerListGauge, prometheus.GaugeValue, cached.(float64))
	} else {
		var avgItems float64
		err := c.db.Model(&struct {
			GroceryListID *uint
		}{}).
			Select("COALESCE(AVG(item_count), 0)").
			Joins("JOIN (SELECT grocery_list_id, COUNT(*) as item_count FROM grocery_entries WHERE grocery_list_id IS NOT NULL GROUP BY grocery_list_id) as counts ON counts.grocery_list_id = grocery_list_id").
			Scan(&avgItems).Error

		if err != nil {
			return
		}

		metricsCache.Set(avgItemsKey, avgItems, cache.DefaultExpiration)
		ch <- prometheus.MustNewConstMetric(c.avgItemsPerListGauge, prometheus.GaugeValue, avgItems)
	}
}

// SetDB sets the database connection for on-demand metrics
func SetDB(database *gorm.DB) {
	db = database

	// If metrics are already initialized, register the collector
	if hasInit.Load() && db != nil {
		registry.MustRegister(NewOnDemandCollector(db))
	}
}

// PrometheusHandler returns the Prometheus metrics endpoint
func PrometheusHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Use our custom registry
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

// UpdateDiscordServers updates the Discord servers count
func UpdateDiscordServers(count int) {
	discordServersGauge.Set(float64(count))
}

// IncrementCommandInvocation increments the counter for a specific command
func IncrementCommandInvocation(commandName string) {
	commandInvocationCounter.WithLabelValues(commandName).Inc()
}

// InitMetrics initializes and registers all metrics
func InitMetrics() {
	if hasInit.Load() {
		return
	}
	hasInit.Store(true)

	// Register the metrics with our custom registry
	registry.MustRegister(discordServersGauge)
	registry.MustRegister(commandInvocationCounter)

	// Register the on-demand collector if we have a database connection
	if db != nil {
		registry.MustRegister(NewOnDemandCollector(db))
	}
}
