package datasource

import (
	"testing"

	"github.com/mr-karan/logchef/internal/clickhouse"
	"github.com/mr-karan/logchef/pkg/models"
)

func TestBoundedViewActivityOptIn(t *testing.T) {
	t.Setenv("LOGCHEF_BOUNDED_ACTIVITY_VIEWS", "observability.logs_applications, observability.logs_platform")
	source := &models.Source{MetaTSField: "Timestamp"}
	source.Connection.Database = "observability"
	source.Connection.TableName = "logs_applications"
	info := &clickhouse.TableInfo{Engine: "View"}
	if !allowsBoundedViewActivity(source, info) {
		t.Fatal("explicit view rejected")
	}
	source.Connection.TableName = "logs_ci"
	if allowsBoundedViewActivity(source, info) {
		t.Fatal("unlisted view admitted")
	}
	source.Connection.TableName = "logs_applications_backup"
	if allowsBoundedViewActivity(source, info) {
		t.Fatal("prefix admitted")
	}
	source.Connection.TableName = "logs_applications"
	source.Connection.Database = "other"
	if allowsBoundedViewActivity(source, info) {
		t.Fatal("other database admitted")
	}
	source.Connection.Database = "observability"
	source.MetaTSField = ""
	if allowsBoundedViewActivity(source, info) {
		t.Fatal("missing timestamp admitted")
	}
	source.MetaTSField = "Timestamp"
	info.Engine = "MergeTree"
	if allowsBoundedViewActivity(source, info) {
		t.Fatal("other engine admitted")
	}
	if allowsBoundedViewActivity(nil, info) || allowsBoundedViewActivity(source, nil) {
		t.Fatal("missing metadata admitted")
	}
	info.Engine = "View"
	t.Setenv("LOGCHEF_BOUNDED_ACTIVITY_VIEWS", "")
	if allowsBoundedViewActivity(source, info) {
		t.Fatal("default safeguard removed")
	}
}
