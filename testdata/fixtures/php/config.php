<?php

$dbHost = getenv('DB_HOST');
$dbPort = $_ENV['DB_PORT'];
$apiKey = $_SERVER['API_KEY'];

$config = [
    'host' => env('DB_HOST'),
    'port' => env('DB_PORT', 5432),
    'key' => $_ENV['API_KEY'],
];

echo "Connecting to $dbHost:$dbPort";
