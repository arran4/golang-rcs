import os

class Revision:
    def __init__(self, rev, date, author, state, next_rev, branches=None):
        self.rev = rev
        self.date = date
        self.author = author
        self.state = state
        self.next_rev = next_rev
        self.branches = branches or []
        self.log = ""
        self.text = ""

    def __str__(self):
        s = f"{self.rev}\n"
        s += f"date\t{self.date};\tauthor {self.author};\tstate {self.state};\n"
        s += "branches"
        for b in self.branches:
            s += f"\n\t{b}"
        s += ";\n"
        s += f"next\t{self.next_rev};\n"
        return s

    def content_str(self):
        s = f"{self.rev}\n"
        log_content = self.log.replace("@", "@@")
        s += f"log\n@{log_content}@\n"
        text_content = self.text.replace("@", "@@")
        s += f"text\n@{text_content}@\n"
        return s

class RCSFile:
    def __init__(self):
        self.head = "1.1"
        self.branch = None
        self.access = []
        self.symbols = {}
        self.locks = []
        self.strict = False
        self.integrity = None
        self.comment = "# "
        self.expand = None
        self.desc = ""
        self.revisions = []

    def __str__(self):
        s = f"head\t{self.head};\n"
        if self.branch:
            s += f"branch\t{self.branch};\n"
        s += "access"
        if self.access:
            s += " " + " ".join(self.access)
        s += ";\n"
        s += "symbols"
        if self.symbols:
            # Sort symbols by key for deterministic output
            # Join with space instead of newline to satisfy parser
            syms = []
            for k in sorted(self.symbols.keys()):
                syms.append(f"{k}:{self.symbols[k]}")
            s += "\n\t" + " ".join(syms)
        s += ";\n"
        s += "locks"
        for l in self.locks:
            s += f"\n\t{l}"
        s += ";\n"
        if self.strict:
            s += "strict;\n"
        if self.integrity:
             integrity_esc = self.integrity.replace("@", "@@")
             s += f"integrity\t@{integrity_esc}@;\n"

        comment_esc = self.comment.replace("@", "@@")
        s += f"comment\t@{comment_esc}@;\n"

        if self.expand:
            expand_esc = self.expand.replace("@", "@@")
            s += f"expand\t@{expand_esc}@;\n"
        s += "\n\n"

        for rev in self.revisions:
            s += str(rev) + "\n"
        s += "\n"

        desc_esc = self.desc.replace("@", "@@")
        s += f"desc\n@{desc_esc}@\n"

        for rev in self.revisions:
            s += "\n\n" + rev.content_str()

        return s

def save_file(filename, content):
    filepath = os.path.join("testdata/generated", filename)
    with open(filepath, "w") as f:
        f.write(content)
    print(f"Generated {filepath}")

def generate_branches():
    f = RCSFile()
    f.head = "1.2"
    f.desc = "File with branches"

    r2 = Revision("1.2", "2023.01.01.00.00.00", "user", "Exp", "1.1", branches=["1.2.1.1"])
    r2.log = "Second revision"
    r2.text = "Line 1\nLine 2\n"

    r1 = Revision("1.1", "2022.01.01.00.00.00", "user", "Exp", "")
    r1.log = "Initial revision"
    r1.text = "Line 1\n"

    r211 = Revision("1.2.1.1", "2023.02.01.00.00.00", "user", "Exp", "")
    r211.log = "Branch revision"
    r211.text = "Line 1\nLine 2\nBranch Line\n"

    f.revisions = [r2, r1, r211]
    save_file("branches.v", str(f))

def generate_access_symbols():
    f = RCSFile()
    f.head = "1.2"
    f.access = ["alice", "bob"]
    f.symbols = {"v1_0": "1.1", "v2_0": "1.2", "beta": "1.2.1.1"}
    f.locks = ["alice:1.2"]
    f.strict = True
    f.desc = "Access and Symbols"

    r2 = Revision("1.2", "2023.01.01.00.00.00", "alice", "Exp", "1.1")
    r2.log = "Rev 2"
    r2.text = "Content 2"

    r1 = Revision("1.1", "2022.01.01.00.00.00", "bob", "Exp", "")
    r1.log = "Rev 1"
    r1.text = "Content 1"

    f.revisions = [r2, r1]
    save_file("access_symbols.v", str(f))

def generate_integrity_expand():
    f = RCSFile()
    f.head = "1.1"
    f.integrity = "some_checksum"
    f.expand = "kv"
    f.desc = "Integrity and Expand"

    r1 = Revision("1.1", "2022.01.01.00.00.00", "user", "Exp", "")
    r1.log = "Rev 1"
    r1.text = "$Id$\n"

    f.revisions = [r1]
    save_file("integrity_expand.v", str(f))

def generate_complex_graph():
    f = RCSFile()
    f.head = "1.3"
    f.desc = "Complex Graph"

    r3 = Revision("1.3", "2023.03.01.00.00.00", "user", "Exp", "1.2")
    r3.log = "Main 3"
    r3.text = "Main content 3"

    r2 = Revision("1.2", "2023.02.01.00.00.00", "user", "Exp", "1.1", branches=["1.2.1.1", "1.2.2.1"])
    r2.log = "Main 2"
    r2.text = "Main content 2"

    r1 = Revision("1.1", "2023.01.01.00.00.00", "user", "Exp", "")
    r1.log = "Main 1"
    r1.text = "Main content 1"

    r211 = Revision("1.2.1.1", "2023.02.05.00.00.00", "dev1", "Exp", "", branches=[])
    r211.next_rev = "1.2.1.2"
    r211.log = "Branch 1.1"
    r211.text = "Branch content 1.1"

    r212 = Revision("1.2.1.2", "2023.02.06.00.00.00", "dev1", "Exp", "")
    r212.log = "Branch 1.2"
    r212.text = "Branch content 1.2"

    r221 = Revision("1.2.2.1", "2023.02.07.00.00.00", "dev2", "Exp", "")
    r221.log = "Branch 2.1"
    r221.text = "Branch content 2.1"

    f.revisions = [r3, r2, r1, r211, r212, r221]
    save_file("complex_graph.v", str(f))

def generate_weird_whitespace():
    f = RCSFile()
    f.head = "1.1"
    f.desc = "Weird Whitespace"
    r1 = Revision("1.1", "2022.01.01.00.00.00", "user", "Exp", "")
    r1.log = "Rev 1"
    r1.text = "Content"
    f.revisions = [r1]

    s = str(f)
    s = s.replace("head\t1.1;", "head  1.1 ; ")
    s = s.replace("desc", "desc   ")
    s = s.replace("1.1", " 1.1 ", 1)

    save_file("weird_whitespace.v", s)

def generate_quoted_strings():
    f = RCSFile()
    f.head = "1.1"
    f.desc = "Quoted @ Strings @"

    r1 = Revision("1.1", "2022.01.01.00.00.00", "user", "Exp", "")
    r1.log = "This log has an @ sign."
    r1.text = "void main() {\n\tprintf(\"Hello @ World\");\n}\n"

    f.revisions = [r1]
    save_file("quoted_strings.v", str(f))

def generate_multiline_symbols():
    f = RCSFile()
    f.head = "1.1"
    f.symbols = {"A":"1.1", "B":"1.1"}
    f.desc = "Multiline symbols"
    r1 = Revision("1.1", "2022.01.01.00.00.00", "user", "Exp", "")
    r1.log = "log"
    r1.text = "text"
    f.revisions = [r1]

    s = str(f)
    # Force multi-line symbols
    s = s.replace("A:1.1 B:1.1", "A:1.1\n\tB:1.1")

    save_file("multiline_symbols.v", s)

if __name__ == "__main__":
    generate_branches()
    generate_access_symbols()
    generate_integrity_expand()
    generate_complex_graph()
    generate_weird_whitespace()
    generate_quoted_strings()
    generate_multiline_symbols()
