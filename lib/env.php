<?php
declare(strict_types=1);
/**
 * Configuration comes ONLY from the environment (or an env file outside the web root).
 * Never commit secrets. See docs/SECURITY-DEPLOY.md.
 *
 * Lookup order: real environment (PHP-FPM `env[...]`, systemd, container, Secrets Manager injection)
 * -> file pointed to by BOGOBOOM_ENV_FILE (default /etc/bogoboom/env, KEY=VALUE per line).
 */
function env(string $key, ?string $default = null): ?string
{
    static $file = null;
    $v = getenv($key);
    if ($v !== false && $v !== '') {
        return $v;
    }
    if ($file === null) {
        $file = [];
        $path = getenv('BOGOBOOM_ENV_FILE') ?: '/etc/bogoboom/env';
        if (is_string($path) && is_readable($path)) {
            foreach (file($path, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES) ?: [] as $line) {
                $line = trim($line);
                if ($line === '' || $line[0] === '#' || !str_contains($line, '=')) {
                    continue;
                }
                [$k, $val] = explode('=', $line, 2);
                $file[trim($k)] = trim($val, " \t\"'");
            }
        }
    }
    return ($file[$key] ?? '') !== '' ? $file[$key] : $default;
}

function env_required(string $key): string
{
    $v = env($key);
    if ($v === null) {
        // message names the variable, never a value
        throw new RuntimeException("Missing required configuration: {$key}");
    }
    return $v;
}
