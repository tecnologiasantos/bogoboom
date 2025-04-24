require_once '../../vendor/autoload.php';

$provider = $_GET['provider'];
$role = $_GET['role'] ?? 'newcomer';

$config = include '../config/hybridauth.php';

try {
  $hybridauth = new Hybridauth\Hybridauth($config);
  $adapter = $hybridauth->authenticate($provider);
  $userProfile = $adapter->getUserProfile();

  // Save or look up user in DB
  // Redirect to appropriate dashboard based on role
} catch (\Exception $e) {
  echo 'Oops, error: ' . $e->getMessage();
}
