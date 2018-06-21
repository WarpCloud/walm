// Copyright 2016 Transwarp Inc. All rights reserved.

{
  local app = self,

  //#############################
  //##### auxiliary function ####
  //#############################
  // handle string(composed by numbers), e.g. "1234"
  // enhancement for std.parseInt,
  parseInt(str)::
    if std.type(str) == 'string' then std.parseInt(str)
    else if std.type(str) == 'number' then str
    else error ('input expected string(composed by numbers) or number, but got ' + std.type(str)),

  convert_jvm_heapsize(memory)::
    if std.type(memory) == 'number' then std.toString(std.floor(memory * 1024)) + 'm'
    else error ('input expected number, but got ' + std.type(memory)),

  strUpper(str)::
    local toupper(ch) =
      if std.codepoint(ch) >= 97 && std.codepoint(ch) <= 122 then std.char(std.codepoint(ch) - 32) else ch;
    std.join('', std.map(toupper, str)),

  strLower(str)::
    local tolower(ch) = if std.codepoint(ch) >= 65 && std.codepoint(ch) <= 90 then std.char(std.codepoint(ch) + 32) else ch;
    std.join('', std.map(tolower, str)),

  shlex(str)::
    local EOF = std.char(3);

    // add char "\u0003" denote the end of string
    local s = str + EOF;
    // whitespace
    local ws(c) =
      if c == ' ' || c == '\n' || c == '\r' || c == '\t' then true
      else false;

    // 5 special chars: ",',`,{,(
    local sc(c) =
      if c == '"' || c == "'" || c == '`' || c == '{' || c == '(' then true
      else false;

    // all printable characters except : " " (space)
    local pc(c) =
      if std.codepoint(c) >= 33 && std.codepoint(c) <= 126 then true
      else false;

    local aux(state, i, token, retArr) =
      if state == 0 then
        if std.length(token) > 0 then  // all tokens are consumed at state 0
          aux(0, i, '', retArr + [token]) tailstrict
        else if s[i] == EOF then
          aux(1, i + 1, token + s[i], retArr) tailstrict
        else if ws(s[i]) then  // don't eat, cursor move forward
          aux(0, i + 1, token, retArr) tailstrict
        else if pc(s[i]) then
          aux(2, i, token, retArr) tailstrict  // state transition, dont't eat, cursor stay
        else
          error ('ValueError: undefined character...in state 0')
      else if state == 1 then
        if token == EOF then
          if std.length(retArr) < 1 then null
          else retArr
        else error ('error state 1.')
      else if state == 2 then
        if s[i] == EOF then
          aux(0, i, token, retArr) tailstrict
        else if ws(s[i]) then
          aux(0, i + 1, token, retArr) tailstrict
        else if s[i] == '\\' then  // don't eat
          aux(3, i + 1, token, retArr) tailstrict
        else if s[i] == "'" then  // don't eat
          aux(5, i + 1, token, retArr) tailstrict
        else if s[i] == '"' then  // don't eat
          aux(7, i + 1, token, retArr) tailstrict
        else if s[i] == '`' then  // don't eat
          aux(10, i + 1, token, retArr) tailstrict
        else if s[i] == '(' then  // eat
          aux(12, i + 1, token + s[i], retArr) tailstrict
        /*else if s[i] == "[" then # eat
          aux(14, i+1, token+s[i], retArr) tailstrict*/
        else if s[i] == '{' then  // eat
          aux(16, i + 1, token + s[i], retArr) tailstrict
        else if pc(s[i]) then
          aux(2, i + 1, token + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 2')
      else if state == 3 then  // out-degree: 10
        if s[i] == EOF then  // don't eat, cursor stay
          aux(0, i, token, retArr) tailstrict
        else if ws(s[i]) || s[i] == 'n' || s[i] == 'r' || s[i] == 't' then
          aux(0, i + 1, token, retArr) tailstrict  // don't eat, cursor move forward
        else if s[i] == '\\' then  // eat
          aux(3, i + 1, token + s[i], retArr) tailstrict
        else if sc(s[i]) then
          aux(2, i, token, retArr) tailstrict  // transition to state 2, handle special char
        else if pc(s[i]) then
          aux(4, i + 1, token + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 3')
      else if state == 4 then  // out-degree: 5+1err
        if s[i] == EOF then
          aux(0, i, token, retArr) tailstrict  // don't eat, cursor stay
        else if ws(s[i]) then
          aux(0, i + 1, token, retArr) tailstrict  // don't eat, cursor either stay or move forward
        else if s[i] == '\\' then
          aux(3, i + 1, token, retArr) tailstrict  // don't eat, cursor move forward
        else if sc(s[i]) then
          aux(2, i, token, retArr) tailstrict
        else if pc(s[i]) && !sc(s[i]) then
          aux(4, i + 1, token + s[i], retArr) tailstrict  // eat
        else error ('ValueError: undefined character...in state 4')
      else if state == 5 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == "'" then  // don't eat
          aux(2, i + 1, token, retArr) tailstrict
        else if s[i] == '\\' then
          aux(6, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then
          aux(5, i + 1, token + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 5')
      else if state == 6 then
        // 5->6, at state 6 whether to eat
        // the pervious character s[i-1]('\') judged by the current char s[i-1]
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == "'" then  // don't eat s[i-1]('\'), '\' acted as escape char
          aux(2, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then  // eat s[i-1]('\')
          aux(5, i + 1, token + s[i - 1] + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 6')
      else if state == 7 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == '"' then
          aux(2, i + 1, token, retArr) tailstrict
        else if s[i] == '\\' then  // don't eat
          aux(8, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then
          aux(7, i + 1, token + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 7')
      else if state == 8 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == '"' then
          aux(2, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then
          aux(7, i, token + s[i - 1] + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 8')
      else if state == 10 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == '`' then
          aux(2, i + 1, token, retArr) tailstrict
        else if s[i] == '\\' then  // don't eat
          aux(11, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then
          aux(10, i + 1, token + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 10')
      else if state == 11 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == '`' then
          aux(2, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then
          aux(10, i, token + s[i - 1] + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 11')
      else if state == 12 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == ')' then
          aux(2, i + 1, token + s[i], retArr) tailstrict
        else if s[i] == '\\' then  // whether to eat judged by next token, line #154
          aux(13, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then
          aux(12, i + 1, token + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 12')
      else if state == 13 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == ')' then
          aux(2, i + 1, token + s[i], retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then  // line #154
          aux(12, i, token + s[i - 1] + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 13')
      else if state == 16 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == '}' then
          aux(2, i + 1, token + s[i], retArr) tailstrict
        else if s[i] == '\\' then  // whether to eat judged by next token, line #172
          aux(17, i + 1, token, retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then
          aux(16, i + 1, token + s[i], retArr) tailstrict
        else error ('ValueError: undefined character...in state 16')
      else if state == 17 then
        if s[i] == EOF then
          aux(18, i, token, retArr) tailstrict
        else if s[i] == '}' then
          aux(2, i + 1, token + s[i], retArr) tailstrict
        else if ws(s[i]) || pc(s[i]) then  // line #172
          aux(16, i, token + s[i - 1] + s[i], retArr) tailstrict
        else error ('error state 17.')
      else if state == 18 then
        error ('ValueError: No closing quotation...')
      else error ('StateError: undefined state ?...')
    ;
    aux(0, 0, '', []),

  //#############################
  //#####   SVC methods #########
  //#############################

  //#############################
  //#####    RC methods #########
  //#############################

  rc_env(config, default)::
    // step 0: type check
    if std.type(config) != 'array' then
      error ('std.filterMap first param must be array, got ' + std.type(config))
    else if std.type(default) != 'array' then
      error ('std.filterMap second param must be array, got ' + std.type(default))
    else
      // step 1: merge & remove duplicates(judge by 'key')
      local config_not_contains(ele) = std.foldl(function(x, y) if y.key == ele.key then false else x, config, true);
      local ans = config + std.filter(config_not_contains, default);

      // step 2: ouput answer, map {key: "", value: ""} to {name: ""， value: ""}
      std.map(function(ele) { name: ele.key, value: ele.value }, ans),


  rc_nodeSelector(config, default)::
    // label_convert
    // TODO step 0: type check
    // TODO step 1: mapping
    local config_map(ele) =
      local ele_split = std.split(ele, '=');
      { [ele_split[0]]: ele_split[1] };

    local config_arr = std.map(config_map, config);
    local config_obj = std.foldl(function(x, y) x + y, config_arr, {});
    if std.length(config_obj) == 0 then null
    else config_obj,
  // config_obj + default,
  // TODO step 2: remove duplicates

  parseNumber(str)::
    local array = std.split(str, '.');
    if std.length(array) == 1 then
      std.parseInt(array[0])
    else if std.length(array) == 2 && std.length(array[1]) >= 1 then
      local digits_after_point = std.length(array[1]);
      std.parseInt(array[0]) + std.parseInt(array[1]) / std.pow(10, digits_after_point)
    else 0,

  diskSizeInByte(size)::
    local isDigitOrPoint(ch) = (std.codepoint(ch) >= 48 && std.codepoint(ch) <= 57) || std.codepoint(ch) == 46;
    local digits(ch) = if isDigitOrPoint(ch) then ch else '';
    local unit(ch) = if isDigitOrPoint(ch) then '' else ch;

    local _size = self.parseNumber(std.join('', std.map(digits, size)));
    local _unit = std.join('', std.map(unit, size));
    local toPower(unit) =
      if std.startsWith(unit, 'G') || std.startsWith(unit, 'g')
      then 1024 * 1024 * 1024
      else if std.startsWith(unit, 'M') || std.startsWith(unit, 'm')
      then 1024 * 1024
      else if std.startsWith(unit, 'K') || std.startsWith(unit, 'k')
      then 1024
      else if std.startsWith(unit, 'B') || std.startsWith(unit, 'b')
      then 1
      else 0;
    std.floor(_size * toPower(_unit)),

  //#############################
  //#####   PD_RC methods #######
  //#############################


}
