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

import org.json.JSONObject;

/**
 * Basic test suite.
 */
public abstract class TestSuite extends Runner
{
	public TestSuite (String inputFile, JSONObject spec)
	{
		super (inputFile, spec);
	}

	protected abstract Object code (Object[] data);

	@Override
	protected String check (Object[] data)
	{
		Object answer = null;
		try
		{
			answer = code (data);
		}
		catch (Exception exception)
		{
			return String.format ("exception:%s", exception);
		}

		String res = moreCheck (answer, data);
		if (! "passed".equals (res))
		{
			return res;
		}

		return String.format ("checked:%s", answer);
	}

	protected String moreCheck (Object answer, Object[] data)
	{
		return "passed";
	}
}
