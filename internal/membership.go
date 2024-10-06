package internal

// Use bloom filter technique for membership query
// This will be the frontier to check whether URL was created before
// if it returns false => 100% false
// if it returns true => maybe true (i.e., false negative can occur)
