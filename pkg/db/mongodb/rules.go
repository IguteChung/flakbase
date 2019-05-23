package mongodb

// canInsert decides whether the given collection and document path is valid with given rule.
func (c *client) canInsert(coll, id string) bool {
	if c.rules == nil {
		// allowed if no rules given.
		return true
	}

	collRule := c.rules.Child(coll)
	docRule := collRule.Child(id)
	return collRule.VariableKey() != "" && docRule.ContainsKey("$other")
}
