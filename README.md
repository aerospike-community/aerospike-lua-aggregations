# `AggAPI.lua`

This UDF can help you with aggregations on the Aerospike database. You need to register this lua UDF on the server using your client's API, and then use aggregations to call it with required parameters.

## How to setup?

You need to register the provided `aggAPI.lua` file as a UDF in your database. Here an example in Go (Note that for the sake of conciseness, the errors are not checked in this example):

```go
luaFile, _ := ioutil.ReadFile("aggAPI.lua")
regTask, _ := client.RegisterUDF(nil, luaFile, "aggAPI.lua", aero.LUA)
// wait until UDF is created on the server.
_ <-regTask.OnComplete()
```

## How does it work?

The Lua streaming UDF will use the argument you pass to it in its calculations by `eval`ing the arguments, and then using them in its logic to calculate and filter the records mostly on the server-side. A last pass will occur on client-side and the results will return in the following format:

```json
{
	[
	  "8de6a795aaf29f2a7dad71c6631a1efc": {
	    "aggs": {
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
	    "aggs": {
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
	]
}
```

The client does not calculate average values, but that can be accomplished as the last step on the client.

The `key` value is the hash used to group the results for reduction. The `aggs` key returns the aggregate values, while the `rec` key returns the bins which were passed as `raw_field`.

Keep in mind that the values are limited to the size of Lua's value size, which is 51 bits of significant integer values.

## How can I calculate a sum?

To calculate the equivalent of the following SQL:

```sql
select sum(salary) from employees
```

call the `select_agg_records` function with following argument:

```json
{
  "field_aliases":         {
    "sum(salary)": "result = rec['salary'] or 0",
  },
  "aggregate_fields": {
    "sum(salary)": "sum",
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
  "filter":    "if rec['age'] ~= nil and rec['age'] > 25 then select_rec = true end",
  "field_aliases":         {
    "sum(salary)": "result =  rec['salary'] or 0",
  },
  "aggregate_fields": {
    "sum(salary)": "sum",
  },
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
  "filter":    "if rec['age'] ~= nil and rec['age'] > 25 then select_rec = true end",
  "field_aliases":         {
    "sum(salary)": "result = rec['salary'] or 0",
  },
  "raw_fields": {
    "age":  "age",
    "name": "name",
  },
  "aggregate_fields": {
    "sum(salary)": "sum",
  },
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
  "filter":    "if rec['age'] ~= nil and rec['age'] >25 then select_rec = true end",
  "field_aliases":         {
    "min(salary)": "result = rec['salary']",
    "max(salary)": "result = rec['salary']",
    "sum(salary)": "result = rec['salary']",
  },
  "raw_fields": {
    "age": "age",
  },
  "aggregate_fields": {
    "min(salary)": "min",
    "max(salary)": "max",
    "sum(salary)": "sum",
  },
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
  "filter":    "if rec['age'] ~= nil and rec['age'] > 25 then select_rec = true end",
  "field_aliases":         {
    "sum(salary * 2)": "result = rec['salary'] * 2",
    "min(salary)":     "result = rec['salary']",
    "max(salary)":     "result = rec['salary']",
  },
  "raw_fields": {
    "age": "age",
  },
  "aggregate_fields": {
    "sum(salary * 2)": "sum",
    "min(salary)":     "min",
    "max(salary)":     "max",
  },
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
  "raw_fields": {
    "age": "age",
  },
  "group_by_fields": [
    "age",
  ],
}
```

## What are the meaning of the values sent to the Lua UDF?

There are 5 different input that need to be sent to the Lua UDF. Not all are required for every command. These values are:

- `"filter"`: Filter is a lua `if` statement. It can include any valid lua boolean statement. These statements have a generic form of:
	`if <boolean expression> then select_rec = true end`

	If the value of the `select_rec` is `true`, the record will be included in the results. Example:
   `if rec['age'] ~= nil and rec['age'] > 25 then select_rec = true end`

- `"field_aliases"`: `fiald_aliases` is the map of the aliases for complex fields. For example:
    - `"sum(salary * 2)": "result = rec['salary'] * 2"` means a field with the name `sum(salary * 2)` should be calculated from the value of the bin `salary` multiplied by 2. 
    - `"min(salary)":     "result = rec['salary']"`: use the value of the bin `salary` to calculate min(salary)    
    - `"max(salary * 5)":     "result = (rec['salary'] or 0) * 5"`: use the value of the bin `salary` multiplied by 5 to calculate the max value. If the `salary` bin is `null`, 0 will be used as default.

- `"raw_fields"`: Raw fields denote fields in the result which do not require any complex calculation. The map key is the alias, while the map value is the name of the existing bin in the database. Example:
  
  `"raw_fields": {
    "age": "age", "salary_usd" : "salary"
  }`
  
- `"aggregate_fields"`: Defines the aggregate type of complex fields defined in `field_aliases`. The valid values are `count`, `sum`, `min` and `max`. Note that the name of the fields should be exactly the same as the ones declared in `field_aliases`. Example:
	  
  `"aggregate_fields": {
    "sum(salary * 2)": "sum",
    "min(salary)":     "min",
    "max(salary * 5)": "max",
  }`
  
- `"group_by_fields"`: List of field aliases to group the fields. Example:
	`[
	    "age", "salary_udf"
	 ]`
