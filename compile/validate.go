package compile

import (
	"fmt"

	"github.com/faiface/funky/parse/parseinfo"
	"github.com/faiface/funky/types"
	"github.com/faiface/funky/types/typecheck"
)

func (env *Env) Validate() []error {
	var errs []error

	for _, definition := range env.names {
		var err error
		switch definition := definition.(type) {
		case *types.Builtin:
		case *types.Record:
			err = env.validateRecord(definition)
		case *types.Union:
			err = env.validateUnion(definition)
		case *types.Alias:
			err = env.validateAlias(definition)
		default:
			panic("unreachable")
		}
		if err != nil {
			errs = append(errs, err)
		}
	}

	for name, impls := range env.funcs {
	implsLoop:
		for i, imp := range impls {
			// check function type
			freeVars := typecheck.FreeVars(imp.TypeInfo()).InOrder()
			err := env.validateType(freeVars, imp.TypeInfo())
			if err != nil {
				errs = append(errs, err)
				continue
			}

			// check other functions for type collisions
			for _, another := range impls[:i] {
				if typecheck.CheckIfUnify(env.names, imp.TypeInfo(), another.TypeInfo()) {
					errs = append(errs, &Error{
						imp.SourceInfo(),
						fmt.Sprintf(
							"function %s with colliding type exists: %v",
							name,
							another.SourceInfo(),
						),
					})
					continue implsLoop
				}
			}
		}
	}

	return errs
}

func (env *Env) validateType(boundVars []string, typ types.Type) error {
	switch typ := typ.(type) {
	case *types.Var:
		for _, bound := range boundVars {
			if typ.Name == bound {
				return nil
			}
		}
		return &Error{typ.SourceInfo(), fmt.Sprintf("type variable not bound: %s", typ.Name)}

	case *types.Appl:
		if env.names[typ.Name] == nil {
			return &Error{typ.SourceInfo(), fmt.Sprintf("type name does not exist: %s", typ.Name)}
		}
		numArgs := len(typ.Args)
		arity := env.names[typ.Name].Arity()
		if numArgs != arity {
			return &Error{
				typ.SourceInfo(),
				fmt.Sprintf("type %s requires %d arguments, %d given", typ.Name, arity, numArgs),
			}
		}
		for _, arg := range typ.Args {
			err := env.validateType(boundVars, arg)
			if err != nil {
				return err
			}
		}
		return nil

	case *types.Func:
		err := env.validateType(boundVars, typ.From)
		if err != nil {
			return err
		}
		err = env.validateType(boundVars, typ.To)
		if err != nil {
			return err
		}
		return nil
	}

	panic("unreachable")
}

func (env *Env) validateRecord(record *types.Record) error {
	err := validateArgs(record.SourceInfo(), record.Args)
	if err != nil {
		return err
	}

	/*// check if all fields have distinct names
	for i, field1 := range record.Fields {
		for _, field2 := range record.Fields[:i] {
			if field1.Name == field2.Name {
				return &Error{
					field1.SI,
					fmt.Sprintf("another record field has the same name: %v", field2.SI),
				}
			}
		}
	}*/

	// validate field types
	for _, field := range record.Fields {
		err := env.validateType(record.Args, field.Type)
		if err != nil {
			return err
		}
	}

	return nil
}

func (env *Env) validateUnion(union *types.Union) error {
	err := validateArgs(union.SourceInfo(), union.Args)
	if err != nil {
		return err
	}

	// check if all alternatives have distinct names
	for i, alt1 := range union.Alts {
		for _, alt2 := range union.Alts[:i] {
			if alt1.Name == alt2.Name {
				return &Error{
					alt1.SI,
					fmt.Sprintf("another union alternative has the same name: %v", alt2.SI),
				}
			}
		}
	}

	// validate alternative types
	for _, alt := range union.Alts {
		for _, field := range alt.Fields {
			err := env.validateType(union.Args, field)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (env *Env) validateAlias(alias *types.Alias) error {
	err := validateArgs(alias.SourceInfo(), alias.Args)
	if err != nil {
		return err
	}
	err = env.validateType(alias.Args, alias.Type)
	if err != nil {
		return err
	}
	return nil
}

func validateArgs(si *parseinfo.Source, args []string) error {
	for i := range args {
		for j := range args[:i] {
			if args[i] == args[j] {
				return &Error{
					si,
					fmt.Sprintf("duplicate type argument: %v", args[i]),
				}
			}
		}
	}
	return nil
}
