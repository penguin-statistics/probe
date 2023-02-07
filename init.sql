CREATE TABLE probe.bonjours
(
    `id` FixedString(26),
    `created_at` DateTime64(6, 'Etc/UTC') DEFAULT now('Etc/UTC'),
    `version` UInt32,
    `platform` LowCardinality(UInt8),
    `uid` FixedString(32),
    `legacy` Bool
)
ENGINE = MergeTree
PRIMARY KEY id
ORDER BY id;

CREATE TABLE probe.impressions
(
    `id` FixedString(26),
    `bonjour_id` FixedString(26),
    `created_at` DateTime('Etc/UTC') DEFAULT now('Etc/UTC'),
    `path` String
)
ENGINE = MergeTree
PRIMARY KEY id
ORDER BY id
SETTINGS index_granularity = 8192;


CREATE TABLE probe.event_search_result_entered
(
    `id` FixedString(26),
    `bonjour_id` FixedString(26),
    `created_at` DateTime('Etc/UTC') DEFAULT now('Etc/UTC'),
    `query` String,
    `result_position` UInt8,
    `destination` String
)
ENGINE = MergeTree
PRIMARY KEY id
ORDER BY id
SETTINGS index_granularity = 8192;