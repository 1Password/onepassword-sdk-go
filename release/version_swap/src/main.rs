/// Helper script to swap the version and build numbers to the ones located in .version and .version-build

use std::fs;
use std::io::{self, BufRead, Write};

// Function to read the contents of a file into a vector of lines
fn read_file_lines(path: &str) -> io::Result<Vec<String>> {
    let file = fs::File::open(path)?;
    let reader = io::BufReader::new(file);
    reader.lines().collect()
}

// Function to write lines to a file
fn write_file_lines(path: &str, lines: &[String]) -> io::Result<()> {
    let mut file = fs::File::create(path)?;
    for line in lines {
        writeln!(file, "{}", line)?;
    }
    Ok(())
}

// Function to replace the version number and build number in the core.go
fn replace_version_number(lines: &[String], new_version: &str, new_build_number: &str) -> Vec<String> {
    lines
        .iter()
        .map(|line| {
            if line.trim().contains("SDKSemverVersion      =") {
                let parts: Vec<&str> = line.split('"').collect();
                format!("{}\"{}\" # v{}", parts[0], new_build_number, new_version)
            } else {
                line.clone()
            }
        })
        .collect()
}
fn main() -> io::Result<()> {
    // Paths to the original file and the version file
    let core_file_path = "../../../internal/core.go";
    let version_file_path = "../../.version";
    let build_number_file_path = "../../.version-build";

    // Acquire the lines of the core.go file and grab the new version/build numbers
    let core_lines = read_file_lines(core_file_path)?;
    let new_version = read_file_lines(version_file_path)?
        .get(0)
        .expect(".version should contain a version number")
        .trim()
        .to_string();

    let new_build_number = read_file_lines(build_number_file_path)?
        .get(0)
        .expect(".version-build is empty")
        .trim()
        .to_string();

    // Replace version number and build number in core.go
    let updated_core_lines = replace_version_number(&core_lines, &new_version, &new_build_number);
    write_file_lines(core_file_path, &updated_core_lines)?;

    Ok(())
}
