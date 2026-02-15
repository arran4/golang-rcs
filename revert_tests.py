import os
import glob

def revert_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    new_lines = []
    for line in lines:
        if line.startswith('# ci') or line.startswith('# co') or line.startswith('# rcs') or line.startswith('# parse error'):
            new_lines.append(line[2:])
        else:
            new_lines.append(line)

    with open(filepath, 'w') as f:
        f.writelines(new_lines)

for f in glob.glob('testdata/txtar/operations/*.txtar'):
    revert_file(f)
for f in glob.glob('testdata/txtar/*.txtar'):
    revert_file(f)
