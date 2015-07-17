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

import os
import re

'''Fill skeleton files containing placeholders with specified values'''
def fillSkeletons(src, dest, fields):
  for (root, dirs, files) in os.walk(src):
    dirdest = dest + root[len(src):]
    # Creates destination directory if not existing yet
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
