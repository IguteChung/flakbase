package esprima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		content string
		program *Program
	}{
		{"auth != null && auth.provider == 'twitter'",
			&Program{
				Type:       "Program",
				SourceType: "script",
				Body: []*ExpressionStatement{
					&ExpressionStatement{
						Type: "ExpressionStatement",
						Expression: &Expression{
							Type: "LogicalExpression",
							LogicalExpression: &LogicalExpression{
								Operator: "&&",
								Left: &Expression{
									Type: "BinaryExpression",
									BinaryExpression: &BinaryExpression{
										Operator: "!=",
										Left: &Expression{
											Type: "Identifier",
											Identifier: &Identifier{
												Name: "auth",
											},
										},
										Right: &Expression{
											Type: "Literal",
											Literal: &Literal{
												Value: nil,
												Raw:   "null",
											},
										},
									},
								},
								Right: &Expression{
									Type: "BinaryExpression",
									BinaryExpression: &BinaryExpression{
										Operator: "==",
										Left: &Expression{
											Type: "MemberExpression",
											MemberExpression: &MemberExpression{
												Computed: false,
												Object: &Expression{
													Type: "Identifier",
													Identifier: &Identifier{
														Name: "auth",
													},
												},
												Property: &Expression{
													Type: "Identifier",
													Identifier: &Identifier{
														Name: "provider",
													},
												},
											},
										},
										Right: &Expression{
											Type: "Literal",
											Literal: &Literal{
												Value: "twitter",
												Raw:   "'twitter'",
											},
										},
									},
								},
							},
						},
					},
				},
			}},
		{
			"newData.hasChildren(['name', 'age'])",
			&Program{
				Type:       "Program",
				SourceType: "script",
				Body: []*ExpressionStatement{
					&ExpressionStatement{
						Type: "ExpressionStatement",
						Expression: &Expression{
							Type: "CallExpression",
							CallExpression: &CallExpression{
								Callee: &Expression{
									Type: "MemberExpression",
									MemberExpression: &MemberExpression{
										Computed: false,
										Object: &Expression{
											Type: "Identifier",
											Identifier: &Identifier{
												Name: "newData",
											},
										},
										Property: &Expression{
											Type: "Identifier",
											Identifier: &Identifier{
												Name: "hasChildren",
											},
										},
									},
								},
								Arguments: []*Expression{
									&Expression{
										Type: "ArrayExpression",
										ArrayExpression: &ArrayExpression{
											Elements: []*Expression{
												&Expression{
													Type: "Literal",
													Literal: &Literal{
														Value: "name",
														Raw:   "'name'",
													},
												},
												&Expression{
													Type: "Literal",
													Literal: &Literal{
														Value: "age",
														Raw:   "'age'",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"newData.val().matches(/^foo/)",
			&Program{
				Type:       "Program",
				SourceType: "script",
				Body: []*ExpressionStatement{
					&ExpressionStatement{
						Type: "ExpressionStatement",
						Expression: &Expression{
							Type: "CallExpression",
							CallExpression: &CallExpression{
								Callee: &Expression{
									Type: "MemberExpression",
									MemberExpression: &MemberExpression{
										Computed: false,
										Object: &Expression{
											Type: "CallExpression",
											CallExpression: &CallExpression{
												Callee: &Expression{
													Type: "MemberExpression",
													MemberExpression: &MemberExpression{
														Computed: false,
														Object: &Expression{
															Type: "Identifier",
															Identifier: &Identifier{
																Name: "newData",
															},
														},
														Property: &Expression{
															Type: "Identifier",
															Identifier: &Identifier{
																Name: "val",
															},
														},
													},
												},
												Arguments: []*Expression{},
											},
										},
										Property: &Expression{
											Type: "Identifier",
											Identifier: &Identifier{
												Name: "matches",
											},
										},
									},
								},
								Arguments: []*Expression{
									&Expression{
										Type: "Literal",
										Literal: &Literal{
											Value: "/^foo/",
											Raw:   "/^foo/",
											Regex: &Regex{
												Pattern: "^foo",
												Flags:   "",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		p, err := Parse(tc.content)
		assert.NoError(t, err)
		assert.EqualValues(t, tc.program, p)
	}
}
