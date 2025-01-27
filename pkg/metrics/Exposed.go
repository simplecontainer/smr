package metrics

var DatabaseSet = NewCounter("database_set_total", "Total database set operations", []string{})
var DatabaseGet = NewCounter("database_get_total", "Total database get operations", []string{})
var DatabasePropose = NewCounter("database_propose_total", "Total database set operations", []string{})
var DatabaseGetKeysPrefix = NewCounter("database_get_prefix_total", "Total database set operations", []string{})
var DatabaseRemove = NewCounter("database_remove_total", "Total database set operations", []string{})
