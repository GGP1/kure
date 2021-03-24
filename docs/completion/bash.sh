# bash completion for kure                                 -*- shell-script -*-

__kure_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__kure_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__kure_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__kure_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__kure_handle_go_custom_completion()
{
    __kure_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly kure allows to handle aliases
    args=("${words[@]:1}")
    requestComp="${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __kure_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __kure_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __kure_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __kure_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __kure_debug "${FUNCNAME[0]}: the completions are: ${out[*]}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __kure_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __kure_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __kure_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out[*]}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __kure_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subDir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out[0]}")
        if [ -n "$subdir" ]; then
            __kure_debug "Listing directories in $subdir"
            __kure_handle_subdirs_in_dir_flag "$subdir"
        else
            __kure_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out[*]}" -- "$cur")
    fi
}

__kure_handle_reply()
{
    __kure_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __kure_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __kure_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __kure_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
		if declare -F __kure_custom_func >/dev/null; then
			# try command name qualified custom func
			__kure_custom_func
		else
			# otherwise fall back to unqualified for compatibility
			declare -F __custom_func >/dev/null && __custom_func
		fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__kure_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__kure_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__kure_handle_flag()
{
    __kure_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __kure_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __kure_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __kure_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __kure_contains_word "${words[c]}" "${two_word_flags[@]}"; then
			  __kure_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__kure_handle_noun()
{
    __kure_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __kure_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __kure_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__kure_handle_command()
{
    __kure_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_kure_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __kure_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__kure_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __kure_handle_reply
        return
    fi
    __kure_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __kure_handle_flag
    elif __kure_contains_word "${words[c]}" "${commands[@]}"; then
        __kure_handle_command
    elif [[ $c -eq 0 ]]; then
        __kure_handle_command
    elif __kure_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __kure_handle_command
        else
            __kure_handle_noun
        fi
    else
        __kure_handle_noun
    fi
    __kure_handle_word
}

_kure_2fa_add()
{
    last_command="kure_2fa_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--digits=")
    two_word_flags+=("--digits")
    two_word_flags+=("-d")
    local_nonpersistent_flags+=("--digits")
    local_nonpersistent_flags+=("--digits=")
    local_nonpersistent_flags+=("-d")
    flags+=("--url")
    flags+=("-u")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_2fa_rm()
{
    last_command="kure_2fa_rm"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_2fa()
{
    last_command="kure_2fa"

    command_aliases=()

    commands=()
    commands+=("add")
    commands+=("rm")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--copy")
    flags+=("-c")
    local_nonpersistent_flags+=("--copy")
    local_nonpersistent_flags+=("-c")
    flags+=("--info")
    flags+=("-i")
    local_nonpersistent_flags+=("--info")
    local_nonpersistent_flags+=("-i")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    local_nonpersistent_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_add_phrase()
{
    last_command="kure_add_phrase"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-e")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    local_nonpersistent_flags+=("-e")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--length=")
    two_word_flags+=("--length")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--length")
    local_nonpersistent_flags+=("--length=")
    local_nonpersistent_flags+=("-l")
    flags+=("--list=")
    two_word_flags+=("--list")
    two_word_flags+=("-L")
    local_nonpersistent_flags+=("--list")
    local_nonpersistent_flags+=("--list=")
    local_nonpersistent_flags+=("-L")
    flags+=("--separator=")
    two_word_flags+=("--separator")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--separator")
    local_nonpersistent_flags+=("--separator=")
    local_nonpersistent_flags+=("-s")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_add()
{
    last_command="kure_add"

    command_aliases=()

    commands=()
    commands+=("phrase")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("p")
        aliashash["p"]="phrase"
        command_aliases+=("passphrase")
        aliashash["passphrase"]="phrase"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--custom")
    flags+=("-c")
    local_nonpersistent_flags+=("--custom")
    local_nonpersistent_flags+=("-c")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-e")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    local_nonpersistent_flags+=("-e")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--length=")
    two_word_flags+=("--length")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--length")
    local_nonpersistent_flags+=("--length=")
    local_nonpersistent_flags+=("-l")
    flags+=("--levels=")
    two_word_flags+=("--levels")
    two_word_flags+=("-L")
    local_nonpersistent_flags+=("--levels")
    local_nonpersistent_flags+=("--levels=")
    local_nonpersistent_flags+=("-L")
    flags+=("--repeat")
    flags+=("-r")
    local_nonpersistent_flags+=("--repeat")
    local_nonpersistent_flags+=("-r")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_backup()
{
    last_command="kure_backup"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--http")
    local_nonpersistent_flags+=("--http")
    flags+=("--path=")
    two_word_flags+=("--path")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_card_add()
{
    last_command="kure_card_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_card_copy()
{
    last_command="kure_card_copy"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--cvc")
    flags+=("-c")
    local_nonpersistent_flags+=("--cvc")
    local_nonpersistent_flags+=("-c")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    local_nonpersistent_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_card_edit()
{
    last_command="kure_card_edit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--it")
    flags+=("-i")
    local_nonpersistent_flags+=("--it")
    local_nonpersistent_flags+=("-i")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_card_ls()
{
    last_command="kure_card_ls"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--filter")
    flags+=("-f")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("-f")
    flags+=("--qr")
    flags+=("-q")
    local_nonpersistent_flags+=("--qr")
    local_nonpersistent_flags+=("-q")
    flags+=("--show")
    flags+=("-s")
    local_nonpersistent_flags+=("--show")
    local_nonpersistent_flags+=("-s")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_card_rm()
{
    last_command="kure_card_rm"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dir")
    flags+=("-d")
    local_nonpersistent_flags+=("--dir")
    local_nonpersistent_flags+=("-d")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_card()
{
    last_command="kure_card"

    command_aliases=()

    commands=()
    commands+=("add")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("create")
        aliashash["create"]="add"
        command_aliases+=("new")
        aliashash["new"]="add"
    fi
    commands+=("copy")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cp")
        aliashash["cp"]="copy"
    fi
    commands+=("edit")
    commands+=("ls")
    commands+=("rm")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_clear()
{
    last_command="kure_clear"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--clipboard")
    flags+=("-c")
    local_nonpersistent_flags+=("--clipboard")
    local_nonpersistent_flags+=("-c")
    flags+=("--terminal")
    flags+=("-t")
    local_nonpersistent_flags+=("--terminal")
    local_nonpersistent_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_config_argon2_test()
{
    last_command="kure_config_argon2_test"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--iterations=")
    two_word_flags+=("--iterations")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--iterations")
    local_nonpersistent_flags+=("--iterations=")
    local_nonpersistent_flags+=("-i")
    flags+=("--memory=")
    two_word_flags+=("--memory")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--memory")
    local_nonpersistent_flags+=("--memory=")
    local_nonpersistent_flags+=("-m")
    flags+=("--threads=")
    two_word_flags+=("--threads")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--threads")
    local_nonpersistent_flags+=("--threads=")
    local_nonpersistent_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_config_argon2()
{
    last_command="kure_config_argon2"

    command_aliases=()

    commands=()
    commands+=("test")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_config_create()
{
    last_command="kure_config_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--path=")
    two_word_flags+=("--path")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    local_nonpersistent_flags+=("-p")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_config_edit()
{
    last_command="kure_config_edit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_config()
{
    last_command="kure_config"

    command_aliases=()

    commands=()
    commands+=("argon2")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("argon")
        aliashash["argon"]="argon2"
    fi
    commands+=("create")
    commands+=("edit")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_copy()
{
    last_command="kure_copy"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    local_nonpersistent_flags+=("-t")
    flags+=("--username")
    flags+=("-u")
    local_nonpersistent_flags+=("--username")
    local_nonpersistent_flags+=("-u")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_edit()
{
    last_command="kure_edit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--it")
    flags+=("-i")
    local_nonpersistent_flags+=("--it")
    local_nonpersistent_flags+=("-i")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_export()
{
    last_command="kure_export"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--path=")
    two_word_flags+=("--path")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    local_nonpersistent_flags+=("-p")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file_add()
{
    last_command="kure_file_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--ignore")
    flags+=("-i")
    local_nonpersistent_flags+=("--ignore")
    local_nonpersistent_flags+=("-i")
    flags+=("--note")
    flags+=("-n")
    local_nonpersistent_flags+=("--note")
    local_nonpersistent_flags+=("-n")
    flags+=("--path=")
    two_word_flags+=("--path")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    local_nonpersistent_flags+=("-p")
    flags+=("--semaphore=")
    two_word_flags+=("--semaphore")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--semaphore")
    local_nonpersistent_flags+=("--semaphore=")
    local_nonpersistent_flags+=("-s")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file_cat()
{
    last_command="kure_file_cat"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--copy")
    flags+=("-c")
    local_nonpersistent_flags+=("--copy")
    local_nonpersistent_flags+=("-c")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file_edit()
{
    last_command="kure_file_edit"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--editor=")
    two_word_flags+=("--editor")
    two_word_flags+=("-e")
    local_nonpersistent_flags+=("--editor")
    local_nonpersistent_flags+=("--editor=")
    local_nonpersistent_flags+=("-e")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file_ls()
{
    last_command="kure_file_ls"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--filter")
    flags+=("-f")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("-f")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file_mv()
{
    last_command="kure_file_mv"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file_rm()
{
    last_command="kure_file_rm"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dir")
    flags+=("-d")
    local_nonpersistent_flags+=("--dir")
    local_nonpersistent_flags+=("-d")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file_touch()
{
    last_command="kure_file_touch"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--overwrite")
    flags+=("-o")
    local_nonpersistent_flags+=("--overwrite")
    local_nonpersistent_flags+=("-o")
    flags+=("--path=")
    two_word_flags+=("--path")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    local_nonpersistent_flags+=("-p")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_file()
{
    last_command="kure_file"

    command_aliases=()

    commands=()
    commands+=("add")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("new")
        aliashash["new"]="add"
    fi
    commands+=("cat")
    commands+=("edit")
    commands+=("ls")
    commands+=("mv")
    commands+=("rm")
    commands+=("touch")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("t")
        aliashash["t"]="touch"
        command_aliases+=("th")
        aliashash["th"]="touch"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_gen_phrase()
{
    last_command="kure_gen_phrase"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--copy")
    flags+=("-c")
    local_nonpersistent_flags+=("--copy")
    local_nonpersistent_flags+=("-c")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-e")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    local_nonpersistent_flags+=("-e")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--length=")
    two_word_flags+=("--length")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--length")
    local_nonpersistent_flags+=("--length=")
    local_nonpersistent_flags+=("-l")
    flags+=("--list=")
    two_word_flags+=("--list")
    two_word_flags+=("-L")
    local_nonpersistent_flags+=("--list")
    local_nonpersistent_flags+=("--list=")
    local_nonpersistent_flags+=("-L")
    flags+=("--mute")
    flags+=("-m")
    local_nonpersistent_flags+=("--mute")
    local_nonpersistent_flags+=("-m")
    flags+=("--qr")
    flags+=("-q")
    local_nonpersistent_flags+=("--qr")
    local_nonpersistent_flags+=("-q")
    flags+=("--separator=")
    two_word_flags+=("--separator")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--separator")
    local_nonpersistent_flags+=("--separator=")
    local_nonpersistent_flags+=("-s")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_gen()
{
    last_command="kure_gen"

    command_aliases=()

    commands=()
    commands+=("phrase")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("p")
        aliashash["p"]="phrase"
        command_aliases+=("passphrase")
        aliashash["passphrase"]="phrase"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--copy")
    flags+=("-c")
    local_nonpersistent_flags+=("--copy")
    local_nonpersistent_flags+=("-c")
    flags+=("--exclude=")
    two_word_flags+=("--exclude")
    two_word_flags+=("-e")
    local_nonpersistent_flags+=("--exclude")
    local_nonpersistent_flags+=("--exclude=")
    local_nonpersistent_flags+=("-e")
    flags+=("--include=")
    two_word_flags+=("--include")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--include")
    local_nonpersistent_flags+=("--include=")
    local_nonpersistent_flags+=("-i")
    flags+=("--length=")
    two_word_flags+=("--length")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--length")
    local_nonpersistent_flags+=("--length=")
    local_nonpersistent_flags+=("-l")
    flags+=("--levels=")
    two_word_flags+=("--levels")
    two_word_flags+=("-L")
    local_nonpersistent_flags+=("--levels")
    local_nonpersistent_flags+=("--levels=")
    local_nonpersistent_flags+=("-L")
    flags+=("--mute")
    flags+=("-m")
    local_nonpersistent_flags+=("--mute")
    local_nonpersistent_flags+=("-m")
    flags+=("--qr")
    flags+=("-q")
    local_nonpersistent_flags+=("--qr")
    local_nonpersistent_flags+=("-q")
    flags+=("--repeat")
    flags+=("-r")
    local_nonpersistent_flags+=("--repeat")
    local_nonpersistent_flags+=("-r")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_import()
{
    last_command="kure_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--erase")
    flags+=("-e")
    local_nonpersistent_flags+=("--erase")
    local_nonpersistent_flags+=("-e")
    flags+=("--path=")
    two_word_flags+=("--path")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    local_nonpersistent_flags+=("-p")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_it()
{
    last_command="kure_it"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_ls()
{
    last_command="kure_ls"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--filter")
    flags+=("-f")
    local_nonpersistent_flags+=("--filter")
    local_nonpersistent_flags+=("-f")
    flags+=("--qr")
    flags+=("-q")
    local_nonpersistent_flags+=("--qr")
    local_nonpersistent_flags+=("-q")
    flags+=("--show")
    flags+=("-s")
    local_nonpersistent_flags+=("--show")
    local_nonpersistent_flags+=("-s")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_restore()
{
    last_command="kure_restore"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_rm()
{
    last_command="kure_rm"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dir")
    flags+=("-d")
    local_nonpersistent_flags+=("--dir")
    local_nonpersistent_flags+=("-d")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_session()
{
    last_command="kure_session"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--prefix=")
    two_word_flags+=("--prefix")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--prefix")
    local_nonpersistent_flags+=("--prefix=")
    local_nonpersistent_flags+=("-p")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    local_nonpersistent_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_stats()
{
    last_command="kure_stats"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kure_root_command()
{
    last_command="kure"

    command_aliases=()

    commands=()
    commands+=("2fa")
    commands+=("add")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("create")
        aliashash["create"]="add"
        command_aliases+=("new")
        aliashash["new"]="add"
    fi
    commands+=("backup")
    commands+=("card")
    commands+=("clear")
    commands+=("config")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cfg")
        aliashash["cfg"]="config"
    fi
    commands+=("copy")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cp")
        aliashash["cp"]="copy"
    fi
    commands+=("edit")
    commands+=("export")
    commands+=("file")
    commands+=("gen")
    commands+=("import")
    commands+=("it")
    commands+=("ls")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("entries")
        aliashash["entries"]="ls"
        command_aliases+=("list")
        aliashash["list"]="ls"
    fi
    commands+=("restore")
    commands+=("rm")
    commands+=("session")
    commands+=("stats")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()


    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_kure()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __kure_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("kure")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function
    local last_command
    local nouns=()

    __kure_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_kure kure
else
    complete -o default -o nospace -F __start_kure kure
fi

# ex: ts=4 sw=4 et filetype=sh
