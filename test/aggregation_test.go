package main_test

import (
	"fmt"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Aggregation Tests", func() {

	Context("Simple Aggregates", func() {

		Context("With no filter", func() {

			It("Should calculate SUM correctly", func() {
				sql := "select sum(age) from test"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"sum(age)": map[string]string{"func": "sum", "expr": "rec['age']"},
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate MIN correctly", func() {
				sql := "select min(age) from test"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"min(age)": map[string]string{"func": "min", "expr": "rec['age']"},
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate MAX correctly", func() {
				sql := "select max(age) from test"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"max(age)": map[string]string{"func": "max", "expr": "rec['age']"},
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate COUNT correctly", func() {
				sql := "select count(age) from test"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"count(age)": map[string]string{"func": "count", "expr": "rec['age'] and 1"},
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate multiple functions correctly", func() {
				sql := "select count(age), min(age*5),max(age+salary), sum(age+1) from test"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"count(age)":      map[string]string{"func": "count", "expr": "rec['age'] and 1"},
						"min(age*5)":      map[string]string{"func": "min", "expr": "rec['age'] * 5"},
						"max(age+salary)": map[string]string{"func": "max", "expr": "rec['age'] + rec['salary']"},
						"sum(age+1)":      map[string]string{"func": "sum", "expr": "rec['age'] + 1"},
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})
		})

		Context("With filter", func() {

			It("Should calculate SUM correctly", func() {
				sql := "select sum(age) from test where age > 20"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"sum(age)": map[string]string{"func": "sum", "expr": "(rec['age'] or 0)"},
					},
					"filter": "rec['age'] ~= nil and rec['age'] > 20",
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate MIN correctly", func() {
				sql := "select min(age) from test where age > 20"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"min(age)": map[string]string{"func": "min", "expr": "rec['age']"},
					},
					"filter": "rec['age'] ~= nil and rec['age'] > 20",
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate MAX correctly", func() {
				sql := "select max(age) from test where age > 20"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"max(age)": map[string]string{"func": "max", "expr": "rec['age']"},
					},
					"filter": "rec['age'] ~= nil and rec['age'] > 20",
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate COUNT correctly", func() {
				sql := "select count(age) from test where age > 20"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"count(age)": map[string]string{"func": "count", "expr": "rec['age'] and 1"},
					},
					"filter": "rec['age'] ~= nil and rec['age'] > 20",
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

			It("Should calculate multiple functions correctly", func() {
				sql := "select count(age), min(age*5),max(age+salary), sum(age+1) from test where age > 20"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"count(age)":      map[string]string{"func": "count", "expr": "rec['age'] and 1"},
						"min(age*5)":      map[string]string{"func": "min", "expr": "rec['age'] * 5"},
						"max(age+salary)": map[string]string{"func": "max", "expr": "rec['age'] + rec['salary']"},
						"sum(age+1)":      map[string]string{"func": "sum", "expr": "rec['age'] + 1"},
					},
					"filter": "rec['age'] ~= nil and rec['age'] > 20",
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror))
			})

		})
	})

	Context("Complex Aggregates with GROUP BY", func() {

		Context("With no filter", func() {

			It("Should calculate SUM correctly", func() {
				sql := "select name, sum(age) from test group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":     "name",
						"sum(age)": map[string]string{"func": "sum", "expr": "rec['age']"},
					},
					// "filter": "rec['age'] == 10",
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "sum(age)"))
			})

			It("Should calculate MIN correctly", func() {
				sql := "select name, min(age) from test group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":     "name",
						"min(age)": map[string]string{"func": "min", "expr": "rec['age']"},
					},
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "min(age)"))
			})

			It("Should calculate MAX correctly", func() {
				sql := "select name, max(age) from test group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":     "name",
						"max(age)": map[string]string{"func": "max", "expr": "rec['age']"},
					},
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "max(age)"))
			})

			It("Should calculate COUNT correctly", func() {
				sql := "select name, count(age) from test group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":       "name",
						"count(age)": map[string]string{"func": "count", "expr": "rec['age'] and 1"},
					},
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "count(age)"))
			})

			It("Should calculate multiple functions correctly", func() {
				sql := "select name, count(age), min(age*5),max(age+salary), sum(age+1) from test group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":            "name",
						"count(age)":      map[string]string{"func": "count", "expr": "rec['age'] and 1"},
						"min(age*5)":      map[string]string{"func": "min", "expr": "rec['age'] * 5"},
						"max(age+salary)": map[string]string{"func": "max", "expr": "rec['age'] + rec['salary']"},
						"sum(age+1)":      map[string]string{"func": "sum", "expr": "rec['age'] + 1"},
					},
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "sum(age+1)", "count(age)", "min(age*5)", "max(age+salary)"))
			})
		})

		Context("With filter", func() {

			It("Should calculate SUM correctly", func() {
				sql := "select name, sum(age) from test where age > 20 group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":     "name",
						"sum(age)": map[string]string{"func": "sum", "expr": "rec['age']"},
					},
					"filter": "rec['age'] > 20",
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "sum(age)"))
			})

			It("Should calculate MIN correctly", func() {
				sql := "select name, min(age) from test where age > 20 group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":     "name",
						"min(age)": map[string]string{"func": "min", "expr": "rec['age']"},
					},
					"filter": "rec['age'] > 20",
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "min(age)"))
			})

			It("Should calculate MAX correctly", func() {
				sql := "select name, max(age) from test where age > 20 group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":     "name",
						"max(age)": map[string]string{"func": "max", "expr": "rec['age']"},
					},
					"filter": "rec['age'] > 20",
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "max(age)"))
			})

			It("Should calculate COUNT correctly", func() {
				sql := "select name, count(age) from test where age > 20 group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":       "name",
						"count(age)": map[string]string{"func": "count", "expr": "rec['age'] and 1"},
					},
					"filter": "rec['age'] > 20",
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "count(age)"))
			})

			It("Should calculate multiple functions correctly", func() {
				sql := "select name, count(age), min(age*5),max(age+salary), sum(age+1) from test where age > 20 group by name"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":            "name",
						"count(age)":      map[string]string{"func": "count", "expr": "rec['age'] and 1"},
						"min(age*5)":      map[string]string{"func": "min", "expr": "rec['age'] * 5"},
						"max(age+salary)": map[string]string{"func": "max", "expr": "rec['age'] + rec['salary']"},
						"sum(age+1)":      map[string]string{"func": "sum", "expr": "rec['age'] + 1"},
					},
					"filter": "rec['age'] > 20",
					"group_by_fields": []string{
						"name",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "sum(age+1)", "count(age)", "min(age*5)", "max(age+salary)"))
			})

			It("Should calculate multiple functions correctly with multiple group by fields", func() {
				sql := "select name, lastname, count(age), min(age*5),max(age+salary), sum(age+1) from test where age > 20 group by name, lastname"
				payload := map[string]interface{}{
					"fields": map[string]interface{}{
						"name":            "name",
						"lastname":        "lastname",
						"count(age)":      map[string]string{"func": "count", "expr": "rec['age'] and 1"},
						"min(age*5)":      map[string]string{"func": "min", "expr": "rec['age'] * 5"},
						"max(age+salary)": map[string]string{"func": "max", "expr": "rec['age'] + rec['salary']"},
						"sum(age+1)":      map[string]string{"func": "sum", "expr": "rec['age'] + 1"},
					},
					"filter": "rec['age'] > 20",
					"group_by_fields": []string{
						"name",
						"lastname",
					},
				}

				sqlr, err := sqlQuery(sqlDB, sql)
				Expect(err).ToNot(HaveOccurred())

				aeror, err := aeroQuery(client, *ns, *set, payload)
				Expect(err).ToNot(HaveOccurred())

				Expect(sqlr).To(MatchQueryResults(aeror, "name", "lastname", "sum(age+1)", "count(age)", "min(age*5)", "max(age+salary)"))
			})
		})

	})

})

func MatchQueryResults(expected interface{}, fieldNames ...string) types.GomegaMatcher {
	return &queryResultMatcher{
		fieldNames: fieldNames,
		expected:   expected,
	}
}

type queryResultMatcher struct {
	fieldNames []string
	expected   interface{}
}

func (matcher *queryResultMatcher) Match(actual interface{}) (success bool, err error) {
	a := actual.([]map[string]interface{})
	b := matcher.expected.([]map[string]interface{})

	if len(matcher.fieldNames) > 0 {
		sort.Slice(a, func(i, j int) bool {
			for _, fname := range matcher.fieldNames {
				switch a[i][fname].(type) {
				case int64:
					if a[i][fname].(int64) == a[j][fname].(int64) {
						continue
					}
					return a[i][fname].(int64) < a[j][fname].(int64)
				case string:
					if a[i][fname].(string) == a[j][fname].(string) {
						continue
					}
					return a[i][fname].(string) < a[j][fname].(string)
				default:
					panic(fmt.Sprintf("Invalid value: %#v", a[i][fname]))
				}
			}
			return false
		})

		sort.Slice(b, func(i, j int) bool {
			for _, fname := range matcher.fieldNames {
				switch b[i][fname].(type) {
				case int64:
					if b[i][fname].(int64) == b[j][fname].(int64) {
						continue
					}
					return b[i][fname].(int64) < b[j][fname].(int64)
				case string:
					if b[i][fname].(string) == b[j][fname].(string) {
						continue
					}
					return b[i][fname].(string) < b[j][fname].(string)
				default:
					panic(fmt.Sprintf("Invalid value: %#v", b[i][fname]))
				}
			}
			return false
		})

	}

	if fmt.Sprintf("%#v", a) == fmt.Sprintf("%#v", b) {
		return true, nil
	}
	return false, nil
}

func (matcher *queryResultMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto match\n\t%#v", actual, matcher.expected)
}

func (matcher *queryResultMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to match\n\t%#v", actual, matcher.expected)
}
