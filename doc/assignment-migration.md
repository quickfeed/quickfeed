# Assignment File Format Migration: YAML to JSON

This document describes the migration from YAML assignment files to JSON assignment files in QuickFeed.

## Overview

QuickFeed now supports both YAML and JSON formats for assignment files:
- `assignment.yml` (legacy YAML format)
- `assignment.yaml` (legacy YAML format)
- `assignment.json` (new JSON format)

## Migration Strategy

### 1. JSON Format Priority

When both YAML and JSON files exist in the same directory, JSON files take precedence over YAML files.

### 2. Conversion Tool

Use the `cm convert-assignments` command to convert existing YAML files to JSON format:

```bash
cm convert-assignments <directory>
```

This command will:
- Recursively search for `assignment.yml` and `assignment.yaml` files
- Convert them to `assignment.json` format
- Preserve all existing field values
- Generate properly formatted JSON files

### 3. Supported Fields

Both YAML and JSON formats support the same fields:

```json
{
  "order": 1,
  "name": "lab1",
  "title": "Assignment Title",
  "deadline": "2025-01-15T23:59:00",
  "scorelimit": 80,
  "autoapprove": false,
  "isgrouplab": false,
  "hoursmin": 10,
  "hoursmax": 15,
  "reviewers": 1,
  "containertimeout": 300
}
```

## Example Migration

### Before (YAML):
```yaml
order: 2
name: "lab2"
title: "Network Programming with REST and gRPC"
deadline: "2025-08-26T23:59:00"
scorelimit: 90
autoapprove: false
isgrouplab: false
hoursmin: 20
hoursmax: 25
```

### After (JSON):
```json
{
  "order": 2,
  "name": "lab2",
  "title": "Network Programming with REST and gRPC",
  "deadline": "2025-08-26T23:59:00",
  "scorelimit": 90,
  "autoapprove": false,
  "isgrouplab": false,
  "hoursmin": 20,
  "hoursmax": 25
}
```

## Benefits

1. **Security**: Removes dependency on YAML parsers that may have security vulnerabilities
2. **Simplicity**: JSON is simpler and more widely supported
3. **Consistency**: Aligns with other JSON-based configuration in the system
4. **Backward Compatibility**: Maintains support for existing YAML files during transition

## Migration Timeline

1. **Phase 1**: Both formats supported (JSON takes precedence)
2. **Phase 2**: Course administrators convert their assignments using the conversion tool
3. **Phase 3**: Eventually, YAML support will be removed

## Usage

### For Course Administrators

1. Run the conversion tool on your tests repository:
   ```bash
   cm convert-assignments /path/to/your/tests/repo
   ```

2. Review the generated JSON files to ensure correctness

3. Commit the JSON files to your repository

4. (Optional) Remove the old YAML files once you're confident the JSON files work correctly

### For Developers

The system automatically handles both formats. When both YAML and JSON files exist for the same assignment, the JSON file will be used.