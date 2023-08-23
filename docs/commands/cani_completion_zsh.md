## cani completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(cani completion zsh)

To load completions for every new session, execute once:

#### Linux:

	cani completion zsh > "${fpath[1]}/_cani"

#### macOS:

	cani completion zsh > $(brew --prefix)/share/zsh/site-functions/_cani

You will need to start a new shell for this setup to take effect.


```
cani completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config string   Path to the configuration file
  -D, --debug           additional debug output
  -v, --verbose         additional verbose output
```

### SEE ALSO

* [cani completion](cani_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 7-Aug-2023