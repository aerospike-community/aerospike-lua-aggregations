local md5 = {
  _VERSION     = "md5.lua 1.1.0",
  _DESCRIPTION = "MD5 computation in Lua (5.1-3, LuaJIT)",
  _URL         = "https://github.com/kikito/md5.lua",
  _LICENSE     = [[
    MIT LICENSE

    Copyright (c) 2013 Enrique García Cota + Adam Baldwin + hanzao + Equi 4 Software

    Permission is hereby granted, free of charge, to any person obtaining a
    copy of this software and associated documentation files (the
    "Software"), to deal in the Software without restriction, including
    without limitation the rights to use, copy, modify, merge, publish,
    distribute, sublicense, and/or sell copies of the Software, and to
    permit persons to whom the Software is furnished to do so, subject to
    the following conditions:

    The above copyright notice and this permission notice shall be included
    in all copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
    OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
    MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
    IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
    CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
    TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
    SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
  ]]
}

-- bit lib implementions

local char, byte, format, rep, sub =
  string.char, string.byte, string.format, string.rep, string.sub
local bit_or, bit_and, bit_not, bit_xor, bit_rshift, bit_lshift

local ok, bit = pcall(require, 'bit')
if ok then
  bit_or, bit_and, bit_not, bit_xor, bit_rshift, bit_lshift = bit.bor, bit.band, bit.bnot, bit.bxor, bit.rshift, bit.lshift
else
  ok, bit = pcall(require, 'bit32')

  if ok then

    bit_not = bit.bnot

    local tobit = function(n)
      return n <= 0x7fffffff and n or -(bit_not(n) + 1)
    end

    local normalize = function(f)
      return function(a,b) return tobit(f(tobit(a), tobit(b))) end
    end

    bit_or, bit_and, bit_xor = normalize(bit.bor), normalize(bit.band), normalize(bit.bxor)
    bit_rshift, bit_lshift = normalize(bit.rshift), normalize(bit.lshift)

  else

    local function tbl2number(tbl)
      local result = 0
      local power = 1
      for i = 1, #tbl do
        result = result + tbl[i] * power
        power = power * 2
      end
      return result
    end

    local function expand(t1, t2)
      local big, small = t1, t2
      if(#big < #small) then
        big, small = small, big
      end
      -- expand small
      for i = #small + 1, #big do
        small[i] = 0
      end
    end

    local to_bits -- needs to be declared before bit_not

    bit_not = function(n)
      local tbl = to_bits(n)
      local size = math.max(#tbl, 32)
      for i = 1, size do
        if(tbl[i] == 1) then
          tbl[i] = 0
        else
          tbl[i] = 1
        end
      end
      return tbl2number(tbl)
    end

    -- defined as local above
    to_bits = function (n)
      if(n < 0) then
        -- negative
        return to_bits(bit_not(math.abs(n)) + 1)
      end
      -- to bits table
      local tbl = {}
      local cnt = 1
      local last
      while n > 0 do
        last      = n % 2
        tbl[cnt]  = last
        n         = (n-last)/2
        cnt       = cnt + 1
      end

      return tbl
    end

    bit_or = function(m, n)
      local tbl_m = to_bits(m)
      local tbl_n = to_bits(n)
      expand(tbl_m, tbl_n)

      local tbl = {}
      for i = 1, #tbl_m do
        if(tbl_m[i]== 0 and tbl_n[i] == 0) then
          tbl[i] = 0
        else
          tbl[i] = 1
        end
      end

      return tbl2number(tbl)
    end

    bit_and = function(m, n)
      local tbl_m = to_bits(m)
      local tbl_n = to_bits(n)
      expand(tbl_m, tbl_n)

      local tbl = {}
      for i = 1, #tbl_m do
        if(tbl_m[i]== 0 or tbl_n[i] == 0) then
          tbl[i] = 0
        else
          tbl[i] = 1
        end
      end

      return tbl2number(tbl)
    end

    bit_xor = function(m, n)
      local tbl_m = to_bits(m)
      local tbl_n = to_bits(n)
      expand(tbl_m, tbl_n)

      local tbl = {}
      for i = 1, #tbl_m do
        if(tbl_m[i] ~= tbl_n[i]) then
          tbl[i] = 1
        else
          tbl[i] = 0
        end
      end

      return tbl2number(tbl)
    end

    bit_rshift = function(n, bits)
      local high_bit = 0
      if(n < 0) then
        -- negative
        n = bit_not(math.abs(n)) + 1
        high_bit = 0x80000000
      end

      local floor = math.floor

      for i=1, bits do
        n = n/2
        n = bit_or(floor(n), high_bit)
      end
      return floor(n)
    end

    bit_lshift = function(n, bits)
      if(n < 0) then
        -- negative
        n = bit_not(math.abs(n)) + 1
      end

      for i=1, bits do
        n = n*2
      end
      return bit_and(n, 0xFFFFFFFF)
    end
  end
end

-- convert little-endian 32-bit int to a 4-char string
local function lei2str(i)
  local f=function (s) return char( bit_and( bit_rshift(i, s), 255)) end
  return f(0)..f(8)..f(16)..f(24)
end

-- convert raw string to big-endian int
local function str2bei(s)
  local v=0
  for i=1, #s do
    v = v * 256 + byte(s, i)
  end
  return v
end

-- convert raw string to little-endian int
local function str2lei(s)
  local v=0
  for i = #s,1,-1 do
    v = v*256 + byte(s, i)
  end
  return v
end

-- cut up a string in little-endian ints of given size
local function cut_le_str(s,...)
  local o, r = 1, {}
  local args = {...}
  for i=1, #args do
    table.insert(r, str2lei(sub(s, o, o + args[i] - 1)))
    o = o + args[i]
  end
  return r
end

local swap = function (w) return str2bei(lei2str(w)) end

-- An MD5 mplementation in Lua, requires bitlib (hacked to use LuaBit from above, ugh)
-- 10/02/2001 jcw@equi4.com

local CONSTS = {
  0xd76aa478, 0xe8c7b756, 0x242070db, 0xc1bdceee,
  0xf57c0faf, 0x4787c62a, 0xa8304613, 0xfd469501,
  0x698098d8, 0x8b44f7af, 0xffff5bb1, 0x895cd7be,
  0x6b901122, 0xfd987193, 0xa679438e, 0x49b40821,
  0xf61e2562, 0xc040b340, 0x265e5a51, 0xe9b6c7aa,
  0xd62f105d, 0x02441453, 0xd8a1e681, 0xe7d3fbc8,
  0x21e1cde6, 0xc33707d6, 0xf4d50d87, 0x455a14ed,
  0xa9e3e905, 0xfcefa3f8, 0x676f02d9, 0x8d2a4c8a,
  0xfffa3942, 0x8771f681, 0x6d9d6122, 0xfde5380c,
  0xa4beea44, 0x4bdecfa9, 0xf6bb4b60, 0xbebfbc70,
  0x289b7ec6, 0xeaa127fa, 0xd4ef3085, 0x04881d05,
  0xd9d4d039, 0xe6db99e5, 0x1fa27cf8, 0xc4ac5665,
  0xf4292244, 0x432aff97, 0xab9423a7, 0xfc93a039,
  0x655b59c3, 0x8f0ccc92, 0xffeff47d, 0x85845dd1,
  0x6fa87e4f, 0xfe2ce6e0, 0xa3014314, 0x4e0811a1,
  0xf7537e82, 0xbd3af235, 0x2ad7d2bb, 0xeb86d391,
  0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476
}

local f=function (x,y,z) return bit_or(bit_and(x,y),bit_and(-x-1,z)) end
local g=function (x,y,z) return bit_or(bit_and(x,z),bit_and(y,-z-1)) end
local h=function (x,y,z) return bit_xor(x,bit_xor(y,z)) end
local i=function (x,y,z) return bit_xor(y,bit_or(x,-z-1)) end
local z=function (ff,a,b,c,d,x,s,ac)
  a=bit_and(a+ff(b,c,d)+x+ac,0xFFFFFFFF)
  -- be *very* careful that left shift does not cause rounding!
  return bit_or(bit_lshift(bit_and(a,bit_rshift(0xFFFFFFFF,s)),s),bit_rshift(a,32-s))+b
end

local function transform(A,B,C,D,X)
  local a,b,c,d=A,B,C,D
  local t=CONSTS

  a=z(f,a,b,c,d,X[ 0], 7,t[ 1])
  d=z(f,d,a,b,c,X[ 1],12,t[ 2])
  c=z(f,c,d,a,b,X[ 2],17,t[ 3])
  b=z(f,b,c,d,a,X[ 3],22,t[ 4])
  a=z(f,a,b,c,d,X[ 4], 7,t[ 5])
  d=z(f,d,a,b,c,X[ 5],12,t[ 6])
  c=z(f,c,d,a,b,X[ 6],17,t[ 7])
  b=z(f,b,c,d,a,X[ 7],22,t[ 8])
  a=z(f,a,b,c,d,X[ 8], 7,t[ 9])
  d=z(f,d,a,b,c,X[ 9],12,t[10])
  c=z(f,c,d,a,b,X[10],17,t[11])
  b=z(f,b,c,d,a,X[11],22,t[12])
  a=z(f,a,b,c,d,X[12], 7,t[13])
  d=z(f,d,a,b,c,X[13],12,t[14])
  c=z(f,c,d,a,b,X[14],17,t[15])
  b=z(f,b,c,d,a,X[15],22,t[16])

  a=z(g,a,b,c,d,X[ 1], 5,t[17])
  d=z(g,d,a,b,c,X[ 6], 9,t[18])
  c=z(g,c,d,a,b,X[11],14,t[19])
  b=z(g,b,c,d,a,X[ 0],20,t[20])
  a=z(g,a,b,c,d,X[ 5], 5,t[21])
  d=z(g,d,a,b,c,X[10], 9,t[22])
  c=z(g,c,d,a,b,X[15],14,t[23])
  b=z(g,b,c,d,a,X[ 4],20,t[24])
  a=z(g,a,b,c,d,X[ 9], 5,t[25])
  d=z(g,d,a,b,c,X[14], 9,t[26])
  c=z(g,c,d,a,b,X[ 3],14,t[27])
  b=z(g,b,c,d,a,X[ 8],20,t[28])
  a=z(g,a,b,c,d,X[13], 5,t[29])
  d=z(g,d,a,b,c,X[ 2], 9,t[30])
  c=z(g,c,d,a,b,X[ 7],14,t[31])
  b=z(g,b,c,d,a,X[12],20,t[32])

  a=z(h,a,b,c,d,X[ 5], 4,t[33])
  d=z(h,d,a,b,c,X[ 8],11,t[34])
  c=z(h,c,d,a,b,X[11],16,t[35])
  b=z(h,b,c,d,a,X[14],23,t[36])
  a=z(h,a,b,c,d,X[ 1], 4,t[37])
  d=z(h,d,a,b,c,X[ 4],11,t[38])
  c=z(h,c,d,a,b,X[ 7],16,t[39])
  b=z(h,b,c,d,a,X[10],23,t[40])
  a=z(h,a,b,c,d,X[13], 4,t[41])
  d=z(h,d,a,b,c,X[ 0],11,t[42])
  c=z(h,c,d,a,b,X[ 3],16,t[43])
  b=z(h,b,c,d,a,X[ 6],23,t[44])
  a=z(h,a,b,c,d,X[ 9], 4,t[45])
  d=z(h,d,a,b,c,X[12],11,t[46])
  c=z(h,c,d,a,b,X[15],16,t[47])
  b=z(h,b,c,d,a,X[ 2],23,t[48])

  a=z(i,a,b,c,d,X[ 0], 6,t[49])
  d=z(i,d,a,b,c,X[ 7],10,t[50])
  c=z(i,c,d,a,b,X[14],15,t[51])
  b=z(i,b,c,d,a,X[ 5],21,t[52])
  a=z(i,a,b,c,d,X[12], 6,t[53])
  d=z(i,d,a,b,c,X[ 3],10,t[54])
  c=z(i,c,d,a,b,X[10],15,t[55])
  b=z(i,b,c,d,a,X[ 1],21,t[56])
  a=z(i,a,b,c,d,X[ 8], 6,t[57])
  d=z(i,d,a,b,c,X[15],10,t[58])
  c=z(i,c,d,a,b,X[ 6],15,t[59])
  b=z(i,b,c,d,a,X[13],21,t[60])
  a=z(i,a,b,c,d,X[ 4], 6,t[61])
  d=z(i,d,a,b,c,X[11],10,t[62])
  c=z(i,c,d,a,b,X[ 2],15,t[63])
  b=z(i,b,c,d,a,X[ 9],21,t[64])

  return bit_and(A+a,0xFFFFFFFF),bit_and(B+b,0xFFFFFFFF),
         bit_and(C+c,0xFFFFFFFF),bit_and(D+d,0xFFFFFFFF)
end

----------------------------------------------------------------

local function md5_update(self, s)
  self.pos = self.pos + #s
  s = self.buf .. s
  for ii = 1, #s - 63, 64 do
    local X = cut_le_str(sub(s,ii,ii+63),4,4,4,4,4,4,4,4,4,4,4,4,4,4,4,4)
    assert(#X == 16)
    X[0] = table.remove(X,1) -- zero based!
    self.a,self.b,self.c,self.d = transform(self.a,self.b,self.c,self.d,X)
  end
  self.buf = sub(s, math.floor(#s/64)*64 + 1, #s)
  return self
end

local function md5_finish(self)
  local msgLen = self.pos
  local padLen = 56 - msgLen % 64

  if msgLen % 64 > 56 then padLen = padLen + 64 end

  if padLen == 0 then padLen = 64 end

  local s = char(128) .. rep(char(0),padLen-1) .. lei2str(bit_and(8*msgLen, 0xFFFFFFFF)) .. lei2str(math.floor(msgLen/0x20000000))
  md5_update(self, s)

  assert(self.pos % 64 == 0)
  return lei2str(self.a) .. lei2str(self.b) .. lei2str(self.c) .. lei2str(self.d)
end

----------------------------------------------------------------

function md5.new()
  return { a = CONSTS[65], b = CONSTS[66], c = CONSTS[67], d = CONSTS[68],
           pos = 0,
           buf = '',
           update = md5_update,
           finish = md5_finish }
end

function md5.tohex(s)
  return format("%08x%08x%08x%08x", str2bei(sub(s, 1, 4)), str2bei(sub(s, 5, 8)), str2bei(sub(s, 9, 12)), str2bei(sub(s, 13, 16)))
end

function md5.sum(s)
  return md5.new():update(s):finish()
end

function md5.sumhexa(s)
  return md5.tohex(md5.sum(s))
end


-----------------------------------------------------------------
-- Aerospike Generic Aggeregation Library
--
-- Copyright 2013-2020 Aerospike Inc.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
-- http:--www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

local function apply_filter_record(rec, filter_func)
  -- if there is no filter, or filter failed to compile: select NO records
  if filter_func == nil then
    return false
  end

  -- if there was a filter specified, and was successfully compiled
  local context = {rec = rec, select_rec = false, string = string}

  -- sandbox the function
  setfenv(filter_func, context)
  filter_func()

  return context.select_rec
end

-----------------------------------------------------------------
-- select_agg_records will return aggregated data
-- the final reduction happens on the client itself
-----------------------------------------------------------------
function select_agg_records(stream, args)
  local aggregate_fields = args["fields"]
  local filter_func_str = args["filter"]
  local group_by_fields = args["group_by_fields"]

  local filter_func = nil
  if filter_func_str ~= nil and #filter_func_str > 0 then
    local eval = loadstring or load
    local err = nil

    filter_func, err = eval("if ("..filter_func_str..") then select_rec = true end")

    if err ~= nil then
      error("Error Parsing Filter: "..err)
    end      
  end

  local aggregate_field_funcs = nil
  if aggregate_fields ~= nil then
    local eval = loadstring or load
    local mapmetadata = getmetatable(map())

    aggregate_field_funcs = {}
    for alias, defs in map.pairs(aggregate_fields) do
      if getmetatable(defs) == mapmetadata then
        local err = nil

        aggregate_field_funcs[alias], err = eval("result = "..defs.expr)
        if err ~= nil then
          error("Error Parsing Expr: "..err)
        end
      else
        if raw_fields == nil then raw_fields = {} end
        raw_fields[alias] = defs
      end
    end
  end


  local function map_aggregates(rec)

    local accu = map()
    local info = map{key = '', rec = map(), agg_results = map({count = 1})}

    if raw_fields ~= nil then
      for alias, v in pairs(raw_fields) do
        local vvv = rec[v]
        if rec[v] ~= nil then
          info.rec[alias] =  vvv
        end
      end
    end

    if aggregate_field_funcs ~= nil then
      for alias, f in pairs(aggregate_field_funcs) do
        local context = {rec = rec, result = nil}

        -- sandbox the function
        setfenv(f, context)
        f()

        local t = type(context.result)
        if t == "number" then
          info.agg_results[alias] = context.result
        elseif t == "nil" then
          -- do nothing; nil is acceptible, but not actionable
        else
          error("Expression for field `"..alias.."` ("..aggregate_fields[alias].expr..") returned a value of type `"..t.."`, instead of number or nil")
        end
      end
    end

    local m = md5.new()
    if group_by_fields ~= nil then
      for v in list.iterator(group_by_fields) do
        local lv = tostring(rec[v] or info.rec[v] or info.agg_results[v])
        -- makes sure sharded values do not concat to the exact same value
        m:update(tostring(#lv))
        m:update(lv)
      end
    end
    local key = md5.tohex(m:finish())

    info.key = key
    accu[key] = info

    return accu
  end

  local function accu_tuples(tuple1, tuple2)
    -- accumulate
    local aggs = map()

    for f, v in map.pairs(tuple1.agg_results) do
      local fn = (aggregate_fields[f] and aggregate_fields[f].func) or ""

      local t1 = tuple1.agg_results[f]
      local t2 = tuple2.agg_results[f]
      aggs[f] = t1

      if fn == "sum" or fn == "count" then
        aggs[f] = v + t2
      elseif fn == "min" then
        if (t2 ~= nil) and (t1 ~= nil) then
          if t2 < t1 then aggs[f] = t2 end
        elseif t2 ~= nil and t1 == nil then 
          aggs[f] = t2
        end
      elseif fn == "max" then
        if (t2 ~= nil) and (t1 ~= nil) then
          if t2 > t1 then aggs[f] = t2 end
        elseif t2 ~= nil and t1 == nil then 
          aggs[f] = t2
        end
      end
    end

    tuple1.agg_results = aggs
    tuple1.agg_results.count = (tuple1.agg_results.count or 0) + (tuple2.agg_results.count or 0)

    return tuple1
  end

  local function reduce_aggregates(accu1, accu2)

    for key, tuple in map.pairs(accu2) do
      if accu1[key] ~= nil then
        -- same key, accumulate
        accu1[key] = accu_tuples(accu1[key], tuple)
      else
        accu1[key] = tuple
      end
    end

    return accu1
  end

  local function filter_records(rec)
    return apply_filter_record(rec, filter_func)
  end

  if filter_func_str ~= nil then
    return stream : filter(filter_records) : map(map_aggregates)  : reduce(reduce_aggregates)
  else
    return stream : map(map_aggregates)  : reduce(reduce_aggregates)
  end
end
