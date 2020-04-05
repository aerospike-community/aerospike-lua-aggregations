/*
 * Copyright 2012-2020 Aerospike, Inc.
 *
 * Portions may be licensed to Aerospike, Inc. under one or more contributor
 * license agreements WHICH ARE COMPATIBLE WITH THE APACHE LICENSE, VERSION 2.0.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */
package com.aerospike.examples;

import com.aerospike.client.AerospikeClient;
import com.aerospike.client.AerospikeException;
import com.aerospike.client.Bin;
import com.aerospike.client.Key;
import com.aerospike.client.Language;
import com.aerospike.client.ResultCode;
import com.aerospike.client.Value;
import com.aerospike.client.policy.Policy;
import com.aerospike.client.query.Filter;
import com.aerospike.client.query.IndexType;
import com.aerospike.client.query.ResultSet;
import com.aerospike.client.query.Statement;
import com.aerospike.client.task.IndexTask;
import com.aerospike.client.task.RegisterTask;
import java.util.*;
import org.json.*;

public class QueryLuaAgg extends Example {

	public QueryLuaAgg(Console console) {
		super(console);
	}

	/**
	 * Query records and calculate count and group by using a user-defined aggregation Package function.
	 *
	 * Functions are part of the Aerospike Lua Aggregations (AggAPI.lua):
	 * https://github.com/aerospike-community/aerospike-lua-aggregations
	 */
	@Override
	public void runExample(AerospikeClient client, Parameters params) throws Exception {
		if (! params.hasUdf) {
			console.info("Query functions are not supported by the connected Aerospike server.");
			return;
		}
		String indexName = "aggindex";
		String keyPrefix = "aggkey";
		String binName = params.getBinName("test_id");
		String binName2 = params.getBinName("group_id");
		int size = 100;

		int test_id = 102;

		// Register the package, once
		register(client, params);

		// Optional, create an index and filter the query using that (once)
		createIndex(client, params, indexName, "test_id");
		writeRecords(client, params, keyPrefix, binName, binName2, size);
		runQuery(client, params, indexName, binName, test_id);
		client.dropIndex(params.policy, params.namespace, params.set, indexName);
	}

	private static Map<String, Object> toMap(JSONObject object) throws JSONException {
		Map<String, Object> map = new HashMap<String, Object>();

		Iterator<String> keysItr = object.keys();
		while(keysItr.hasNext()) {
			String key = keysItr.next();
			Object value = object.get(key);

			if(value instanceof JSONArray) {
				value = toList((JSONArray) value);
			}

			else if(value instanceof JSONObject) {
				value = toMap((JSONObject) value);
			}
			map.put(key, value);
		}
		return map;
	}

	private static List<Object> toList(JSONArray array) throws JSONException {
		List<Object> list = new ArrayList<Object>();
		for(int i = 0; i < array.length(); i++) {
			Object value = array.get(i);
			if(value instanceof JSONArray) {
				value = toList((JSONArray) value);
			}

			else if(value instanceof JSONObject) {
				value = toMap((JSONObject) value);
			}
			list.add(value);
		}
		return list;
	}

	private void register(AerospikeClient client, Parameters params) throws Exception {
		RegisterTask task = client.register(params.policy, "udf/aggAPI.lua", "aggAPI.lua", Language.LUA);
		// Alternately register from resource.
		task.waitTillComplete();
	}

	private void createIndex(
		AerospikeClient client,
		Parameters params,
		String indexName,
		String binName
	) throws Exception {
		console.info("Create index: ns=%s set=%s index=%s bin=%s",
			params.namespace, params.set, indexName, binName);

		Policy policy = new Policy();
		policy.socketTimeout = 0; // Do not timeout on index create.

		try {
			IndexTask task = client.createIndex(policy, params.namespace, params.set, indexName, binName, IndexType.NUMERIC);
			task.waitTillComplete();
		}
		catch (AerospikeException ae) {
			if (ae.getResultCode() != ResultCode.INDEX_ALREADY_EXISTS) {
				throw ae;
			}
		}
	}

	private static int getRandomNumberInRange(int min, int max) {

		Random r = new Random();
		return r.ints(min, (max + 1)).limit(1).findFirst().getAsInt();

	}
	private void writeRecords(
		AerospikeClient client,
		Parameters params,
		String keyPrefix,
		String binName,
		String binName2,
		int size
	) throws Exception {
		for (int i = 1; i <= size; i++) {
			Key key = new Key(params.namespace, params.set, keyPrefix + i);
			Bin bin = new Bin(binName, getRandomNumberInRange(100, 105));
			Bin bin2 = new Bin(binName2, getRandomNumberInRange(1, 3));

			console.info("Put: ns=%s set=%s key=%s bin=%s value=%s bin2=%s value2=%s",
				key.namespace, key.setName, key.userKey, bin.name, bin.value, bin2.name, bin2.value);

			client.put(params.writePolicy, key, bin, bin2);
		}
	}

	private void runQuery(
		AerospikeClient client,
		Parameters params,
		String indexName,
		String binName,
		Integer testId
	) throws Exception {

		/*
{
  "fields": {
    "test_id": "test_id",
    "group_id": "group_id",
    "count(group_id)": {"func": "count", "expr": "result = rec['group_id'] ~= nil and 1"},
    "count(*)":     {"func": "count", "expr": "result = 1"},
  },
  "filter":    "rec['test_id'] ~= nil and rec['test_id'] == %s",
  "group_by_fields": [
    "test_id",
    "group_id",
  ],
}
		*/

		String stringToParse = String.format("{\n" +
				"  \"fields\": {\n" +
				"    \"test_id\": \"test_id\",\n" +
				"    \"group_id\": \"group_id\",\n" +
				"    \"count(group_id)\": {\"func\": \"count\", \"expr\": \"rec['group_id'] ~= nil and 1\"},\n" +
				"    \"count(*)\":     {\"func\": \"count\", \"expr\": \"1\"},\n" +
				"  },\n" +
				"  \"filter\":    \"rec['test_id'] ~= nil and rec['test_id'] == %s\",\n" +
				"  \"group_by_fields\": [\n" +
				"    \"test_id\",\n" +
				"    \"group_id\",\n" +
				"  ],\n" +
				"}", testId.toString());

		JSONObject json = new JSONObject(stringToParse);
		Map functionArgsMap  = toMap(json);

		console.info("Query for: ns=%s set=%s index=%s bin=%s",
			params.namespace, params.set, indexName, binName);
		console.info("Using query: %s",
				functionArgsMap.toString());

		Statement stmt = new Statement();
		stmt.setNamespace(params.namespace);
		stmt.setSetName(params.set);
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
								res.get("group_id"),
								res.get("count(*)"));
					});
				}
			}
		}

		finally {
			rs.close();
		}
	}
}
