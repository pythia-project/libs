// Pythia library for unit testing-based tasks
// Author: Sébastien Combéfis <sebastien@combefis.be>
// 
// Copyright (C) 2019, Computer Science and IT in Education ASBL
// Copyright (C) 2019, ECAM Brussels Engineering School
// 
// This program is free software: you can redistribute it and/or modify
// under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 2 of the License, or
//  (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package org.pythia;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.io.Reader;
import java.io.Writer;

import org.apache.commons.csv.CSVFormat;
import org.apache.commons.csv.CSVParser;
import org.apache.commons.csv.CSVRecord;

import org.json.JSONArray;
import org.json.JSONObject;

/**
 * Basic code runner.
 */
public abstract class Runner
{
	private String inputFile;
	private JSONObject spec;

	protected Runner (String inputFile, JSONObject spec)
	{
		this.inputFile = inputFile;
		this.spec = spec;
	}

	protected String check (Object[] data)
	{
		return String.format ("%s", code (data));
	}

	protected abstract Object code (Object[] data);

	private Object[] parseTestData (CSVRecord data)
	{
		JSONArray params = spec.getJSONArray ("args");
		Object[] args = new Object[params.length()];
		for (int i = 0; i < params.length(); i++)
		{
			args[i] = parse (data.get (i), params.getJSONObject (i).getString ("type"));
		}
		return args;
	}

	private Object parse (String data, String type)
	{
		switch (type)
		{
			case "int":
				return Integer.valueOf (data);
		}
		return null;
	}

	public void run (String dest, String filename) throws IOException
	{
		try
		(
			// Create the results file
			Writer writer = new BufferedWriter (new FileWriter (dest + "/" + filename));

			// Read and run tests
			Reader reader = new BufferedReader (new FileReader (inputFile));
			CSVParser csvParser = new CSVParser (reader, CSVFormat.DEFAULT
				.withDelimiter(';')
				.withQuote('"'));
		)
		{
			for (CSVRecord row : csvParser)
			{
				String res = check (parseTestData (row));
				writer.write (res);
				writer.write ('\n');
			}
		}
	}
}
