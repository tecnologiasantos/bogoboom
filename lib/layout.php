<?php
declare(strict_types=1);
/** Minimal shared page chrome for PHP-rendered pages (all dynamic values must go through h()). */
function page_start(string $title, string $heading = 'Welcome to Bogoboom', string $sub = ''): void
{
    echo '<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">'
        . '<title>' . h($title) . '</title><style>'
        . 'body{font-family:Arial,sans-serif;margin:0;background:#f5f5f5;color:#222}header{background:#004a8f;color:#fff;padding:1rem 2rem;text-align:center}'
        . 'main{padding:2rem;max-width:1100px;margin:0 auto}section{background:#fff;padding:1.5rem 2rem;border-radius:8px;box-shadow:0 4px 8px rgba(0,0,0,.1);margin:1.5rem auto;overflow-x:auto}'
        . 'h3{color:#004a8f}form.grid{display:grid;gap:1rem;max-width:420px;margin:0 auto}label{font-weight:bold;color:#004a8f}'
        . 'input,button{width:100%;padding:.8rem;border-radius:6px;font-size:1rem;border:1px solid #ddd;box-sizing:border-box}'
        . 'button{background:#004a8f;color:#fff;cursor:pointer;border:0}button.danger{background:#c0392b}button.warn{background:#e08a00}'
        . 'table{width:100%;border-collapse:collapse}th,td{padding:.7rem;border:1px solid #ddd;text-align:left}th{background:#004a8f;color:#fff}'
        . '.err{color:#b00020}.ok{color:#146c2e}.actions{display:flex;gap:1rem;justify-content:center;margin-top:1rem}.actions form{margin:0}footer{text-align:center;font-size:.9rem;color:#666;margin:2rem 0}'
        . '</style></head><body><header><h1>' . h($heading) . '</h1>' . ($sub !== '' ? '<p>' . h($sub) . '</p>' : '') . '</header><main>';
}

function page_end(): void
{
    echo '</main><footer>&copy; ' . date('Y') . ' Bogoboom. All rights reserved.</footer></body></html>';
}
