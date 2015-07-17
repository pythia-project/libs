# -*- coding: utf-8 -*-
#
# Pythia library for unit testing-based tasks
# Author: Sébastien Combéfis <sebastien@combefis.be>
#
# Copyright (C) 2015, Computer Science and IT in Education ASBL
# Copyright (C) 2015, École Centrale des Arts et Métiers
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

import csv
import os
import random
import re

'''Fill skeleton files containing placeholders with specified values'''
def fillSkeletons(src, dest, fields):
  for (root, dirs, files) in os.walk(src):
    dirdest = dest + root[len(src):]
    # Create destination directory if not existing yet
    if not os.path.exists(dirdest):
      os.makedirs(dirdest)
    # Handle each file in the src directory
    for filename in files:
      filesrc = '{}/{}'.format(root, filename)
      filedest = '{}/{}'.format(dirdest, filename)
      # Open the file
      with open(filesrc, 'r', encoding='utf-8') as file:
        content = file.read()
      # Replace each placeholder with the specified value
      for field, value in fields.items():
        regex = re.compile('@([^@]*)@{}@([^@]*)@'.format(field))
        for prefix, postfix in set(regex.findall(content)):
          rep = '\n'.join([prefix + v + postfix for v in value.splitlines()])
          content = content.replace('@{}@{}@{}@'.format(prefix, field, postfix), rep)
      # Create the new file
      with open(filedest, 'w', encoding='utf-8') as file:
        file.write(content)

'''Generate input data for random tests'''
def generateRandomTestData(dest, filename, config):
  # Create destination directory if not existing yet
  if not os.path.exists(dest):
    os.makedirs(dest)
  with open('{}/{}'.format(dest, filename), 'w', encoding='utf-8') as file:
    writer = csv.writer(file, delimiter=';', quotechar='"')
    generator = ArrayGenerator([RandomGenerator.build(descr) for descr in config['args']])
    for i in range(config['n']):
      writer.writerow(generator.generate())

class RandomGenerator:
  def generate(self):
    return None

  def build(description):
    m = re.match('^int\((-{0,1}[1-9][0-9]*),(-{0,1}[1-9][0-9]*)\)', description)
    if not m is None:
      return IntRandomGenerator(int(m.group(1)), int(m.group(2)))
    return RandomGenerator()

class ArrayGenerator(RandomGenerator):
  def __init__(self, generators):
    self.generators = generators

  def generate(self):
    return [g.generate() for g in self.generators]

class IntRandomGenerator(RandomGenerator):
  def __init__(self, lowerbound, upperbound):
    self.lowerbound = lowerbound
    self.upperbound = upperbound

  def generate(self):
    return random.randint(self.lowerbound, self.upperbound)

'''Test suite'''
class TestSuite:
  def __init__(self, predefined, inputfile):
    self.predefined = predefined
    self.inputfile = inputfile

  def check(self, data):
    try:
      answer = self.studentCode(data)
    except Exception as e:
      return 'exception:{}'.format(e)
    res = self.moreCheck(answer, data)
    if res != 'passed':
      return res
    return 'checked:{}'.format(answer)

  def studentCode(self, data):
    return None

  def moreCheck(self, answer, data):
    return 'passed'

  def parseTestData(self, data):
    return tuple(data)

  def run(self, resultfile):
    with open(resultfile, 'w', encoding='utf-8') as file:
      # Run predefined tests
      for data in self.predefined:
        res = self.check(self.parseTestData(data))
        file.write('{}\n'.format(res))
    pass
