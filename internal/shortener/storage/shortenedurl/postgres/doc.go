/*
Package postgres defines a shortURL postgres repo.

shortURLSaver, shortURLProvider, shortURLDeleter
defines different types for working separately with postgres repo.

The same pgxpool.Pool is used to create any of them.
*/
package postgres
