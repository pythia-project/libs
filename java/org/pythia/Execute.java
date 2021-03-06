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
import java.io.FileInputStream;
import java.io.FileReader;
import java.io.InputStreamReader;
import java.io.IOException;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;

import java.util.Arrays;
import java.util.List;

import org.json.JSONArray;
import org.json.JSONObject;
import org.pythia.Runner;
import org.pythia.TestSuite;

public class Execute
{
	private static String readAll (String path) throws IOException
	{
		BufferedReader reader = new BufferedReader (new InputStreamReader (new FileInputStream (path), "UTF-8"));
		StringBuilder buff = new StringBuilder();
		String line = null;
		while ((line = reader.readLine()) != null)
		{
			buff.append(line);
		}
		return buff.toString();
	}

	public static void main (String[] args) throws ClassNotFoundException, IOException, NoSuchMethodException
	{
		if (args.length < 1 || ! Arrays.asList("student", "teacher").contains(args[0]))
		{
			System.exit(1);
		}

		// Read the specification of the function.
		JSONObject spec = new JSONObject (readAll ("/task/config/spec.json"));

		// Import the code to execute.
		Class<?> program = ClassLoader.getSystemClassLoader().loadClass ("Program");
		String name = spec.getString ("name");
		JSONArray params = spec.getJSONArray ("args");
		Class<?>[] paramsType = new Class<?>[params.length()];
		for (int i = 0; i < params.length(); i++)
		{
			switch (params.getJSONObject (i).getString ("type"))
			{
				case "int":
					paramsType[i] = int.class;
			}
		}
		Method method = program.getDeclaredMethod (name, paramsType);

		// Create the specific runner for the code to execute.
		String inputFile = "/tmp/work/input/data.csv";
		Runner runner = null;
		String outputFile = null;
		if ("student".equals (args[0]))
		{
			runner = new TestSuite (inputFile, spec) {
				@Override
				protected Object code (Object[] data)
				{
					Object result = null;
					try
					{
						result = method.invoke (null, data);
					}
					catch (IllegalAccessException | InvocationTargetException exception){}
					return result;
				}
			};
			outputFile = "data.res";
		}
		else
		{
			runner = new Runner (inputFile, spec) {
				@Override
				protected Object code (Object[] data)
				{
					Object result = null;
					try
					{
						result = method.invoke (null, data);
					}
					catch (IllegalAccessException | InvocationTargetException exception){}
					return result;
				}
			};
			outputFile = "solution.res";
		}

		runner.run ("/tmp/work/output", outputFile);
	}
}
