# -*- coding: utf-8 -*-
#
# Pythia library for unit testing-based tasks
# Author: Sébastien Combéfis <sebastien@combefis.be>
#
# Copyright (C) 2015-2019, Computer Science and IT in Education ASBL
# Copyright (C) 2015-2016, ECAM Brussels Engineering School
#
# This program is free software: you can redistribute it and/or modify
# under the terms of the GNU General Public License as published by
# the Free Software Foundation, version 2 of the License, or
#  (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
# General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

from abc import ABC, abstractmethod
import csv

class Runner(ABC):
    '''Basic code runner.'''
    def __init__(self, inputfile, spec):
        self.inputfile = inputfile
        self.spec = spec

    def check(self, data):
        return '{}'.format(self.code(data))

    @abstractmethod
    def code(self, data):
        pass

    def parseTestData(self, data):
        return tuple(self._parse(data[i], self.spec['args'][i]['type']) for i in range(len(data)))

    def _parse(self, data, type):
        if type == 'int':
            return int(data)
        if type == 'bool':
            return bool(data)
        if type == 'float':
            return float(data)
        if type in ['enum', 'str']:
            return data
        if type[:2] == '[]':
            if type[2:] == 'int':
                if data == '[]':
                    return []
                return [int(x) for x in data[1:len(data)-1].split(' ')]
        return None

    def run(self, dest, filename):
        # Create the results file
        with open('{}/{}'.format(dest, filename), 'w', encoding='utf-8') as result:
            # Read and run tests
            with open(self.inputfile, 'r', encoding='utf-8') as file:
                reader = csv.reader(file, delimiter=';', quotechar='"')
                for row in reader:
                    res = self.check(self.parseTestData(row))
                    result.write('{}\n'.format(res))


class TestSuite(Runner):
    '''Basic test suite.'''
    def __init__(self, inputfile, spec):
        Runner.__init__(self, inputfile, spec)

    def check(self, data):
        try:
            answer = self.code(data)
        except Exception as e:
            return 'exception:{}'.format(e)
        res = self.moreCheck(answer, data)
        if res != 'passed':
            return res
        return 'checked:{}'.format(answer)

    def moreCheck(self, answer, data):
        return 'passed'
