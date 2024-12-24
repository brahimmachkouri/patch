## The compiled file is way too big, so I made a rust version : [https://github.com/brahimmachkouri/patchit](patchit)

# Patch a file or create a patch 

Applies a patch to a file using a JSON file that contains the name of the file to patch with its checksum, along with the offsets and the byte data to be modified :

```bash
.\patchit.exe mybinary.json
```

Generate a patch (json file) by comparing 2 files :

```bash
.\patchit.exe --source mybinary.orig.exe --modified mybinary.exe --output mybinary.json
```

Example of generated JSON file:
```json
	{
	  "file_name": "mybinary.exe",
	  "checksum": "79935e89d59728ac456b592ca7b4f64dee0f3a7bb10e44e068cf0c635f885735",
	  "patches": [
	    {
	      "offset": 190577,
	      "data": "75"
	    },
	    {
	      "offset": 1139552,
	      "data": "31"
	    },
	    {
	      "offset": 1139553,
	      "data": "c0"
	    },
	    {
	      "offset": 1139554,
	      "data": "c3"
	    }
	  ]
	}
```

Copyright (c) 2024 Brahim Machkouri

This software is provided "as is", without any warranty of any kind, express or implied, including but not limited to the warranties of merchantability and fitness for a particular purpose. In no event shall the author or copyright holders be liable for any damage, whether in an action of contract, tort, or otherwise, arising from the use of this software.
