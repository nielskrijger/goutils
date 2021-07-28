// Package validator implements value validations
//
// Copyright 2014 Roberto Teixeira <robteix@robteix.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
/*
	Validator is a modification of https://github.com/go-validator/validator.

	It contains the following validation rules:
		- gte=4: tests whether a variable value is larger or equal to a given
		  number. For number types, it's a simple greater-than test; for strings
  		  it tests the number of characters whereas for maps and slices it tests
		  the number of items.
		- lte=4: tests whether a variable value is smaller or equal to a given
		  number. For number types, it's a simple lesser-than test; for strings
		  it tests the number of characters whereas for maps and slices it tests
		  the number of items.
		- required: checks whether a variable is non-zero as defined by the
		  golang spec. You're advised not to use this validation for booleans
		  and numbers,
	    - since golang defaults empty numbers to 0 and empty booleans to false.
		- name: string containing unicode letters -,.' and not start or end
		  with a space.
		- az09_: string containing 0-9, A-Z, _ and not start with a _.
		- gender: string either "male", "female" or "genderqueer".
		- isodate: a time.Time where hour, minute, second and millisecond are 0.
		  If value is a string checks if date is in YYYY-MM-DD format.
		- zoneinfo: zoneinfo timestamp, e.g. Europe/Amsterdam.
		- locale: space-separated string of BCP47 language tags.
		- mindate=2006-01-02: time.Time with a minimum date. "now" will use
		  today's date.
		- maxdate=2006-01-02: time.Time with a maximum date "now" will use
		  today's date.
		- url: accepts any url the golang request uri accepts.

	In addition the following tags are aliases:
		- username: "az09_,gte=4,lte=20"
		- birthdate: "isodate,mindate=1900-01-01,maxdate=now"

*/
package validate
