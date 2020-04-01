# `AggAPI.lua`

This UDF will help you with arbitrary aggregations on the Aerospike database. You need to register this lua UDF on the server using your client's API, and then use aggregation API on your chosen Aerospike Client to call it with required parameters.

## Limitations

Aerospike Server supports Lua 5.1, in which all numbers are floats with 51 bits significands (52 with sign bit). This means integers bigger than 2^51 will return invalid values.

## How to setup?

You need to register the provided `aggAPI.lua` file as a UDF in your database. Here an example in Go (Note that for the sake of conciseness, the errors are not checked in this example):

```go
luaFile, _ := ioutil.ReadFile("aggAPI.lua")
regTask, _ := client.RegisterUDF(nil, luaFile, "aggAPI.lua", aero.LUA)
// wait until UDF is created on the server.
_ <-regTask.OnComplete()
```

```java
RegisterTask task = client.register(params.policy, "udf/aggAPI.lua", "aggAPI.lua", Language.LUA);
// Alternately register from resource.
task.waitTillComplete();
```

## How does it work?

The Lua streaming UDF will use the argument you pass to it in its calculations by `eval`ing the arguments, and then using them in its logic to calculate and filter the records mostly on the server-side. A last pass will occur on client-side and the results will return in the following format:

```json
{
  "8de6a795aaf29f2a7dad71c6631a1efc": {
    "agg_results": {
      "count(age)":      3.000000,
      "max(age)":        45.000000,
      "count":           3.000000,
      "sum(age*salary)": 101400,
      "min(age)":        25.000000,
    },
    "key": "8de6a795aaf29f2a7dad71c6631a1efc",
    "rec": {
      "name": "Eva",
    },
  },
  "ed57af7ff6ed54ec8b6b5eec3e2b649a": {
    "agg_results": {
      "count(age)":      1.000000,
      "max(age)":        26.000000,
      "count":           1.000000,
      "sum(age*salary)": 83200,
      "min(age)":        26.000000,
    },
    "key": "ed57af7ff6ed54ec8b6b5eec3e2b649a",
    "rec": {
      "name": "Riley",
    },
  },
}
```

The client does not calculate average values, but that can be accomplished as the last step on the client.

Regardless of the aggregations you have asked, the count of final records will always be returned in `aggs.count`. Try to avoid this name in your requests.

The `key` value is the hash used to group the results for reduction. The `aggs` key returns the aggregate values, while the `rec` key returns the bins which were passed as `raw_field`.

Keep in mind that the values are limited to the size of Lua's value size, which is 51 bits of significant integer values.

Example in Go:
```go
stm := aero.NewStatement(nsName, setName)

functionArgsMap := map[string]interface{}{
  "fields": map[string]interface{}{
    "name": "name",
    "max(age)":        map[string]string{"func": "max", "expr": "rec['age'] ~= nil and rec['age']"},
    "count(age)":      map[string]string{"func": "count", "expr": "( rec['age'] ) ~= nil and 1"},
    "min(age)":        map[string]string{"func": "min", "expr": "rec['age'] ~= nil and rec['age']"},
    "sum(age*salary)": map[string]string{"func": "sum", "expr": " (rec['age']  or 0) * (rec['salary'] or 0)"},
  },
  "filter": "rec['age'] ~= nil and rec['age'] >5 ",
  "group_by_fields": []string{
    "name",
  },
}

recordset, err := client.QueryAggregate(nil, stm, "aggAPI", "select_agg_records", aero.NewValue(functionArgsMap))
defer recordset.Close()

if err != nil {
  return err
}

for result := range recordset.Results() {
  if result.Err != nil {
    return result.Err
  }

  pp.Println(result.Record.Bins["SUCCESS"])
}
```

Example in Java:
```java
String stringToParse = String.format("{\n" +
    "  \"fields\":         {\n" +
    "    \"test_id\": \"test_id\",\n" +
    "    \"state\": \"state\",\n" +
    "    \"count(state)\": {\"func\":\"count\", \"expr\": \"rec['state'] ~= nil and 1\"},\n" +
    "  },\n" +
    "  \"filter\":    \"rec['test_id'] ~= nil and rec['test_id'] == %s\",\n" +
    "  \"group_by_fields\": [\n" +
    "    \"test_id\",\n" +
    "    \"state\",\n" +
    "  ],\n" +
    "}", testId.toString());

JSONObject json = new JSONObject(stringToParse);
Map functionArgsMap  = toMap(json);

Statement stmt = new Statement();
stmt.setNamespace(namespace);
stmt.setSetName(set);
stmt.setBinNames(binName);

// Optional filter via Index
stmt.setFilter(Filter.equal(binName, testId));
stmt.setIndexName(indexName);

ResultSet rs = client.queryAggregate(null, stmt,"aggAPI","select_agg_records", Value.get(functionArgsMap));

try {
  if (rs.next()) {
    Object obj = rs.getObject();

    if (obj instanceof Map<?, ?>) {
      Map<String, Map> map = (Map<String, Map>) obj;

      for (Map<?,?> x: map.values()){
        console.warn("res: " +
            ((Map<?,?>)x.get("rec")).get("test_id") + " " +
            ((Map<?,?>)x.get("rec")).get("state") + " " +
            ((Map<?,?>)x.get("agg_results")).get("count"));
      }
    }
  }
}

finally {
  rs.close();
}

```

## How can I calculate a sum?

To calculate the equivalent of the following SQL:

```sql
select sum(salary) from employees
```

call the `select_agg_records` function with following argument:

```json
{
  "fields":         {
    "sum(salary)": {"func": "sum" , "expr": "rec['salary'] or 0"},
  },
}
```

## Adding conditions

To add a condition like:

```sql
select sum(salary) from employees where age > 25
```

provide the following arguments:

```json
{
  "fields":         {
    "sum(salary)": {"func": "sum" , "expr": "rec['salary'] or 0"},
  },
  "filter":    "rec['age'] ~= nil and rec['age'] > 25",
}
```

## Adding `Group BY`

To add other fields to the query like:

```sql
select name, age, sum(salary) from employees where age > 25 group by name, age
```

provide the following arguments:

```json
{
  "fields":         {
    "age":  "age",
    "name": "name",
    "sum(salary)": {"func": "sum" , "expr": "rec['salary'] or 0"},
  },
  "filter":    "rec['age'] ~= nil and rec['age'] > 25",
  "group_by_fields": [
    "age",
    "name",
  ],
}
```

keep in mind that the UDF logic does not validate if your provided `group by` arguments are valid for the logic of the equivalent SQL command.

## Adding `min` and `max` functions

To add other fields to the query like:

```sql
select age, min(salary), max(salary), sum(salary) from employees where age > 25 group by age
```

provide the following arguments:

```json
{
  "fields":         {
    "age": "age",
    "min(salary)": {"func": "min" , "expr": "rec['salary']"},
    "max(salary)": {"func": "max" , "expr": "rec['salary']"},
    "sum(salary)": {"func": "sum" , "expr": "rec['salary']"},
  },
  "filter":    "rec['age'] ~= nil and rec['age'] > 25",
  "group_by_fields": [
    "age",
  ],
}
```

## Can I do more complex statements in the functions and filters?

YES! The execute the equivalent of the following SQL command:

```sql
select age, min(salary), max(salary), sum(salary * 2) from employees where age > 25 group by age
```

provide the following arguments:

```json
{
  "fields":         {
    "age": "age",
    "sum(salary * 2)": {"func": "sum" , "expr": "rec['salary'] * 2"},
    "min(salary)":     {"func": "min" , "expr": "rec['salary']"},
    "max(salary)":     {"func": "max" , "expr": "rec['salary']"},
  },
  "filter":    "rec['age'] ~= nil and rec['age'] > 25",
  "group_by_fields": [
    "age",
  ],
}
```

## How can I do DISTINCT queries?

In case you would want to return the following SQL statement:

```sql
select distinct age from employees
```

since you can rewrite the above query as:

```sql
select age from employees group by age
```

then the parameters sent to the UDF would be:

```json
{
  "fields": {
    "age": "age",
  },
  "group_by_fields": [
    "age",
  ],
}
```

## What are the meaning of the values sent to the Lua UDF?

There are 5 different input that need to be sent to the Lua UDF. Not all are required for every command. These values are:

- `"raw_fields"`: Raw fields denote fields in the result which do not require any complex calculation. The map key is the alias, while the map value is the name of the existing bin in the database. Example:
  
  `"raw_fields": {
    "age": "age", 
    "salary_usd" : "salary"
  }`

- `"fields"`: `fields` is the map of the aliases for complex and calculated fields. For example:
    - `"sum(salary * 2)": {"func": "sum", "expr": "rec['salary'] * 2"}` means a field with the name `sum(salary * 2)` should be calculated from the value of the bin `salary` multiplied by 2. 
    - `"min(salary)":     {"func": "sum", "expr": "rec['salary']"}`: use the value of the bin `salary` to calculate `min(salary)`
    - `"max(salary * 5)": {"func": "sum", "expr": "(rec['salary'] or 0) * 5"}`: use the value of the bin `salary` multiplied by 5 to calculate the max value. If the `salary` bin is `null`, 0 will be used as default value.
  
- `"filter"`: Filter is a lua boolean statement.

  If the value of the `statement` is `true`, the record will be included in the results. Example:
   `rec['age'] ~= nil and rec['age'] > 25`
  
- `"group_by_fields"`: List of field aliases to group the fields. Example:
	`[
	    "age", "salary_udf"
	 ]`
