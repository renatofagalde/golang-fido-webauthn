// Package ptr atraves da funcao Of, devolve um ponteiro para uma copia de v.
// v é um parametro por valor, a funcao recebe uma copia, e o endereco retornardo aponta para essa copia,
// nao para o original do caller
// É uma copia e pega o endereco numa expressao só
//
// ptr é abreviacao de ponteiro, convenção antiga em C,
// a comunidade em GO herdou
package ptr

func Of[X any](v X) *X {
	return &v
}
