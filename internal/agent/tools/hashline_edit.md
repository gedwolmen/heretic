Use this tool to edit a file when you have read it with the `view` tool
(which produces line#ID tagged output). The edit is rejected if the file
changed between read and edit.

For each line you want to edit, you need:
- `target_line`: the exact line text (without the #ID suffix)
- `line_hash`: the 4-character #ID hash from the read output

This is the safest way to edit a file with the LLM, because stale edits
are detected and rejected before they can corrupt the file.
