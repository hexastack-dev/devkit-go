// This package provide additional features to standard errors package without
// changing method signature and behaviour with non-intrusive way. More specifically
// this package add ability to annotate error with caller information, see New and
// Errorf method for more detail. Ideally, you should only annotate the root error
// either newly created error or received error from other libs, you should avoid
// annotate the same error chain twice, this way you can avoid cluttered the error
// logs.
//
// Other method are simply calling standard errors method of the same name.
package errors
