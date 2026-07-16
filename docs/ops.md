<h1>abort</h1>

  AbortOp can be used to abort execution


<h3>Arguments</h3>
  - <code>message</code>
  The reason of abort

<h3>Examples</h3>
  - Abort execution unconditionally
  ```yaml
  abort:
    message: Pipeline aborted
  ```

  - Abort execution when data in the specific path, in the tree is not equal to some value
  ```yaml
  abort:
    message: Pipeline aborted
  when: '{{ not (eq 1502 .result.code) }}'
  ```

<h1>call</h1>

  Call is used to invoke named action, previously defined using DefineOp


<h3>Arguments</h3>
  - <code>args</code>
  Arguments to be passed to callable.
Leaf values are recursively templated just before call is executed.

  - <code>argsFrom</code>
  When specified, this is a path within the global data where to take arguments from.
This takes precedence over "args" which are ignored.

  - <code>argsPath</code>
  ArgsPath is optional path within the global data where arguments are stored prior to execution. When omitted, then default value of "args" is assumed. Note that passing arguments to nested callable is only possible if path is different, otherwise inner's arguments will overwrite outer's one. Template is accepted as possible value.

  - <code>name</code>
  Name is name of callable previously registered using DefineOp.
Attempt to use name that was not registered will result in error.
Template is supported

<h1>define</h1>

  DefineOp can be used to define the ActionSpec and later recall it by name via CallOp.
Attempt to define name that was defined before will result in an error.


<h3>Arguments</h3>
  - <code>action</code>
  %!s(<nil>)

  - <code>name</code>
  Name that will be used for registration

<h1>env</h1>

  This op is used to import OS environment variables into data


<h3>Arguments</h3>
  - <code>exclude</code>
  Optional regexp which defines what to exclude.
Only item names NOT matching this regexp are added into data document.
Exclusion is considered after inclusion regexp is processed.

  - <code>include</code>
  Optional regexp which defines what to include.
Only item names matching this regexp are added into data document.

  - <code>path</code>
  Optional path within data tree under which "Env" container will be put. When omitted, then "Env" goes to root of data.

<h1>exec</h1>

  Executes external program using OS's exec


<h3>Arguments</h3>
  - <code>args</code>
  Optional arguments for program

  - <code>dir</code>
  Program's working directory

  - <code>program</code>
  Program to execute

  - <code>saveExitCodeTo</code>
  Path within the global data where to set exit code.

  - <code>stderr</code>
  Path to file where program's stderr will be written upon completion.
Any error occurred during write will result in error.

  - <code>stdout</code>
  Path to file where program's stdout will be written upon completion.
Any error occurred during write will result in error.

  - <code>validExitCodes</code>
  List of exit codes that are assumed to be valid

<h1>export</h1>

  Exports data in desired format from the data tree


<h3>Arguments</h3>
  - <code>file</code>
  File to export data onto

  - <code>format</code>
  Format of output file

  - <code>path</code>
  Path within data tree pointing to dom.Node to export. Empty path denotes whole document.
If path does not resolve, then empty document will be exported.
If output format is "text" then path must point to leaf.
Any other output format must point to dom.Container.
If neither of these conditions are met, then it is considered as if path does not resolve at all.

<h1>ext</h1>

  ExtOp invokes extension function, previously registered with runtime


<h3>Arguments</h3>
  - <code>args</code>
  holds arguments to be passed to function

  - <code>function</code>
  Name of the function that was registered with the Executor

<h1>import</h1>

  Imports data from file into data tree


<h3>Arguments</h3>
  - <code>file</code>
  File to read data from

  - <code>mode</code>
  How to parse the file before the import takes place

  - <code>path</code>
  Path at which to import the data.

  - <code>xml</code>
  XML/HTML loading options

<h1>log</h1>

  LogOp just logs message to logger


<h3>Arguments</h3>
  - <code>message</code>
  Message to log

<h1>patch</h1>

  PatchOp performs RFC6902-style patch on global data document.


<h3>Arguments</h3>
  - <code>from</code>
  Path used as a source with Copy and Move operations

  - <code>op</code>
  Op is RFC6902 operation

  - <code>path</code>
  Path is used as general path for every operation

  - <code>value</code>
  Value to be used for op. This takes precedence over ValueFrom.

  - <code>valueFrom</code>
  Allow a value to be read from data tree at given path.
 Only considered when Value is not specified

<h3>Examples</h3>
  - Delete first item from list "employees" under "hr" element in the root of data
  ```yaml
  op: remove
  path: /hr/employees/0
  ```

  - Add new item at position 2, to the list "employees" under "hr" element in the root of data
  ```yaml
  op: add
  path: /hr/employees/2
  value:
    department: Management
    name: Bob
  ```

<h1>set</h1>

  Sets the data into data tree


<h3>Arguments</h3>
  - <code>data</code>
  Arbitrary data to put into data tree

  - <code>path</code>
  Path at which to put data.
If omitted, then data are merged into root of document

  - <code>render</code>
  Flag indicating use of templating. When true, data are passed through template engine

  - <code>strategy</code>
  Strategy defines how that are handled when conflict during set/add of data occur.

<h1>templateFile</h1>

  TemplateFileOp can be used to render template from file and write result to output.


<h3>Arguments</h3>
  - <code>file</code>
  Path to file with template

  - <code>output</code>
  Output is path to output file

  - <code>path</code>
  Path is path within the global data where data are read from (must be container). Those data are set to template engine.When omitted, then root of global data is assumed

<h1>template</h1>

  TemplateOp can be used to render value from data at runtime


<h3>Arguments</h3>
  - <code>parseAs</code>
  How to treat rendered text after template engine completes successfully.
t's responsibility of template to produce source that is parseable by chosen mode

  - <code>path</code>
  Path within global data tree where to set result at

  - <code>template</code>
  Template to render

  - <code>trim</code>
  When true, whitespace is trimmed off the value
