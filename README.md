# `AggAPI.lua`

This UDF will help you with arbitrary aggregations on the Aerospike database. You need to register this lua UDF on the server using your client's API, and then use aggregation API on your chosen Aerospike Client to call it with required parameters.

## Limitations

Aerospike Server supports Lua 5.1, in which all numbers are floats with 51 bits significands (52 with sign bit). This means integers bigger than 2^51 will return invalid values.

## How to setup?

You need to register the provided `aggAPI.lua` file as a UDF in your database. 

Here an example in Go (note that for the sake of conciseness, the errors are not checked in this example):
```go
luaFile, _ := ioutil.ReadFile("aggAPI.lua")
regTask, _ := client.RegisterUDF(nil, luaFile, "aggAPI.lua", aero.LUA)
// wait until UDF is created on the server.
_ <-regTask.OnComplete()
```

Using Java to register the module:
```java
RegisterTask task = client.register(params.policy, "udf/aggAPI.lua", "aggAPI.lua", Language.LUA);
// Alternately register from resource.
task.waitTillComplete();
```

Using AQL to register the module:
```
aql> register module 'aggAPI.lua'
```

## How does it work?

The Lua streaming UDF will use the argument you pass to it in its calculations by `eval`ing the arguments, and then using them in its logic to calculate and filter the records mostly on the server-side. 
```json
    "fields": {
        "name":            "name",
        "max(age)":        {"func": "max", "expr": "rec['age'] ~= nil and rec['age']"},
        "count(age)":      {"func": "count", "expr": "( rec['age'] ) ~= nil and 1"},
        "min(age)":        {"func": "min", "expr": "rec['age'] ~= nil and rec['age']"},
        "sum(age*salary)": {"func": "sum", "expr": " (rec['age']  or 0) * (rec['salary'] or 0)"},
    },
    "filter": "rec['age'] ~= nil and rec['age'] >5 ",
    "group_by_fields": {
      "name",
    }
```

A last pass will occur on client-side and the results will return in the following format:
```json
{
  "8de6a795aaf29f2a7dad71c6631a1efc": {
    "count(age)":      3.000000,
    "max(age)":        45.000000,
    "count":           3.000000,
    "sum(age*salary)": 101400,
    "min(age)":        25.000000,
    "name":            "Eva",
  },
  "ed57af7ff6ed54ec8b6b5eec3e2b649a": {
    "count(age)":      1.000000,
    "max(age)":        26.000000,
    "count":           1.000000,
    "sum(age*salary)": 83200,
    "min(age)":        26.000000,
    "name":            "Riley",
  },
}
```
The `key` is a hash used to group the results for reduction. The value is a map of the returned fields. In the map, the key is the alias of the field.

Regardless of the aggregations used, a count of records in the group will always be returned in `count`. Avoid using this name in your requests.

The client does not calculate average values, but that can be accomplished as the last step at the client (sum and count).

Keep in mind that the values are limited to the size of Lua's value size, which is 51 bits of significant integer values.

## What are the meaning of the values sent to the Lua UDF?

There are 3 different input that need to be sent to the Lua UDF. Not all are required for every command. These values are:

- `"fields"`: Choosing the fields to return - this is the equivalent of the `select` part of the query.
    - Fields which do not require any complex calculation: the map key is the alias, while the map value is the name of the existing bin in the database.  
    Example:
      ```json
      fields": {
        "age": "age",  
        "salary_usd" : "salary"  
      }
      ```
    - Fields which are calculated (apply an aggregate function on): the map key is the aliases, and the value is a map of the function and its calculation.  
      Example:
      ```json
      "fields": {
        "sum(salary * 2)": {"func": "sum", "expr": "rec['salary'] * 2"}, 
        "min(salary)":     {"func": "sum", "expr": "rec['salary']"},
        "max(salary * 5)": {"func": "sum", "expr": "(rec['salary'] or 0) * 5"}
      }
      ```
        - `"sum(salary * 2)": {"func": "sum", "expr": "rec['salary'] * 2"}` means a field with the name `sum(salary * 2)` should be calculated from the value of the bin `salary` multiplied by 2. 
        - `"min(salary)":     {"func": "sum", "expr": "rec['salary']"}`: use the value of the bin `salary` to calculate `min(salary)`
        - `"max(salary * 5)": {"func": "sum", "expr": "(rec['salary'] or 0) * 5"}`: use the value of the bin `salary` multiplied by 5 to calculate the max value. If the `salary` bin is `null`, 0 will be used as default value.
  
- `"filter"`: Filter is a lua boolean statement to filter records - this is the equivalent of a `where` in a query.  
  If the value of the `statement` is `true`, the record will be included in the results.
  
  Example:   
   `"filter": "rec['age'] ~= nil and rec['age'] > 25"`
  
- `"group_by_fields"`: List of field aliases to group the records by - this is the equivalent of `group by` in a query.   
Example:
    ```json
    "group_by_fields": [
        "age", "salary_udf"
     ]
    ```

## Example: Building a Query
### How can I calculate a sum?

To calculate the equivalent of the following SQL:

```sql
select sum(salary) from employees
```

We would call the `select_agg_records` function with following argument:

```json
{
  "fields":         {
    "sum(salary)": {"func": "sum" , "expr": "rec['salary'] or 0"},
  },
}
```
Which means use the value of the bin `salary`. If the `salary` bin is `null` or does not exist, 0 will be used as default value.
We name this calculation `sum(salary)`.

### Adding a conditions

To add a filtering condition:

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

Which means check if the `age` bin exists, and if it does, check if the age is over 25.

### Adding a `Group BY`

To create a group by aggregation, we would need to add the fields both to the `fields` part and to the group by.

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
**Please note: the UDF logic does not validate if your provided `group by` arguments are valid for the logic of the equivalent SQL command.**

### Adding `min` and `max` functions

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

### Can I do more complex statements in the functions and filters?

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

## How can I do `DISTINCT` queries?

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

## Code Examples
###Example in Go:
```go
stm := aero.NewStatement(nsName, setName)

functionArgsMap := map[string]interface{}{
  "fields": map[string]interface{}{
    "name":            "name",
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

### Example in Java:
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

        map.values().forEach(res -> {
            console.info(res.toString());

            console.info("res: %s, %s => %s",
                    res.get("test_id"),
                    res.get("state"),
                    res.get("count(*)"));
      }
    }
  }
}

finally {
  rs.close();
}
```
