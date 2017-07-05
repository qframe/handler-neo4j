package handler_neo4j

import (
	"fmt"
	"github.com/zpatrick/go-config"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/qframe/types/inventory"
	"strings"
)


type Neo4jConnector struct {
	conn bolt.Conn
}

func NewNeo4jConnector(cfg *config.Config) (Neo4jConnector, error) {
	nc := Neo4jConnector{}
	host, _ := cfg.StringOr("host", "localhost")
	port, _ := cfg.StringOr("port", "7687")
	addr := fmt.Sprintf("bolt://%s:%s", host, port)
	err := nc.Connect(addr)
	return nc, err
}

func (nc *Neo4jConnector) Connect(addr string) (err error) {
	nc.conn, err = bolt.NewDriver().OpenNeo(addr)
	return
}

func (nc *Neo4jConnector) Close() {
	nc.conn.Close()
}

func (nc *Neo4jConnector) UpsertNode(node interface{}) (err error) {
	switch node.(type) {
	case string:
		q := `MERGE (o:Node {name: {name}})
  			ON CREATE SET o.created = timestamp(),o.seen = timestamp()
  			ON MATCH SET  o.seen = timestamp()`
		_, err = nc.conn.ExecNeo(q, map[string]interface{}{"name": node})
	}
	return
}

func (nc *Neo4jConnector) ConnectNode(subject, action, object string) (err error) {
	params := map[string]interface{}{"subject": subject, "action": action, "object": object}
	q := fmt.Sprintf(`
	MATCH  (s:Node {name: {subject}})
	MATCH  (o:Node {name: {object}})
	MERGE (s)-[:%s]->(o)
	`, strings.ToUpper(action))
	_, err = nc.conn.ExecNeo(q, params)
	return
}

func (nc *Neo4jConnector) ConnectBase(inv qtypes_inventory.Base) (err error) {
	err = nc.UpsertNode(inv.Subject)
	if err != nil {
		return
	}
	err = nc.UpsertNode(inv.Object)
	if err != nil {
		return
	}
	return nc.ConnectNode(inv.Subject, inv.Action, inv.Object)
}