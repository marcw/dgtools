# dgtools

A command line utility to work with the Discogs data dumps. Download dumps, convert them to parquet, import them into a PostgreSQL database, get the XML structure of a dump, etc...

## Usage

```
dgtools [global options] command [command options] [arguments...]
```

### Global Options

- `--discogs-bucket` - The URL of the Discogs data dumps (default: "https://discogs-data-dumps.s3.us-west-2.amazonaws.com")

## Commands

### dump

Work with Discogs data dump files.

#### dump list

List the files in the Discogs data dumps.

```
dgtools dump list [options]
```

**Options:**
- `--year` - Filter by year
- `--month` - Filter by month  
- `--type` - Filter by data type
- `--no-table` - Don't print the table (output filenames only)

#### dump structure

Dump the structure of an XML file.

```
dgtools dump structure <file> [options]
```

**Arguments:**
- `file` - The file to dump the structure of

**Options:**
- `--stop-after X` - Stops analysis after X records


#### dump download

Download a Discogs data dump.

```
dgtools dump download [options] <name>
```

**Arguments:**
- `name` - The file to download


**Options:**
- `--out-dir` - The output directory (default: ".")
- `--overwrite` - Force the download even if the file already exists
- `--checksum` - Check the checksum of the file after downloading (default: true)

#### dump convert

Convert a dump to a different format

```
dgtools dump convert <name> --out <name> [options]
```

**Arguments:**
- `name` - The file to convert

**Options:**
- `--out` - The output file
- `--stop-after X` - Stop conversion after X records


### db

Work with a database.

**Options:**
- `--database-url` - The URL of the database to connect to (default: "postgres://$USER@localhost:5432/dgtools", can be set via DATABASE_URL environment variable)

#### db prepare

Prepare the database for import by running migrations.

```
dgtools db prepare
```

#### db import

Import data from a dump file to the database.

```
dgtools db import <file>
```

**Arguments:**
- `file` - The file to import the data from

#### db nuke

Nuke the database by rolling back all migrations.

```
dgtools db nuke
```

## LICENSE

Please see [LICENSE.md](LICENSE.md)
