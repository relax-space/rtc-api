-- --------------------------------------------------------
-- 主机:                           127.0.0.1
-- 服务器版本:                        5.7.22 - MySQL Community Server (GPL)
-- 服务器操作系统:                      Linux
-- HeidiSQL 版本:                  9.5.0.5295
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;


-- 导出 rtc 的数据库结构
CREATE DATABASE IF NOT EXISTS `rtc` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `rtc`;

-- 导出  表 rtc.db_account 结构
CREATE TABLE IF NOT EXISTS `db_account` (
  `tenant_name` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `host` varchar(255) DEFAULT NULL,
  `port` int(11) DEFAULT NULL,
  `user` varchar(255) DEFAULT NULL,
  `pwd` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 正在导出表  rtc.db_account 的数据：~3 rows (大约)
/*!40000 ALTER TABLE `db_account` DISABLE KEYS */;
INSERT INTO `db_account` (`tenant_name`, `name`, `host`, `port`, `user`, `pwd`) VALUES
	('pangpang', 'mysql', '127.0.0.1', 3308, 'root', '******'),
	('pangpang', 'sqlserver', '10.202.101.230', 1433, 'sa', '******'),
	('test', 'mysql', '127.0.0.1', 3311, 'root', '1234');
/*!40000 ALTER TABLE `db_account` ENABLE KEYS */;

-- 导出  表 rtc.image_account 结构
CREATE TABLE IF NOT EXISTS `image_account` (
  `registry` varchar(255) DEFAULT NULL,
  `login_name` varchar(255) DEFAULT NULL,
  `pwd` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 正在导出表  rtc.image_account 的数据：~0 rows (大约)
/*!40000 ALTER TABLE `image_account` DISABLE KEYS */;
/*!40000 ALTER TABLE `image_account` ENABLE KEYS */;

-- 导出  表 rtc.namespace 结构
CREATE TABLE IF NOT EXISTS `namespace` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tenant_name` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- 正在导出表  rtc.namespace 的数据：~1 rows (大约)
/*!40000 ALTER TABLE `namespace` DISABLE KEYS */;
INSERT INTO `namespace` (`id`, `tenant_name`, `name`) VALUES
	(1, 'pangpang', 'pangpang-brand');
/*!40000 ALTER TABLE `namespace` ENABLE KEYS */;

-- 导出  表 rtc.project 结构
CREATE TABLE IF NOT EXISTS `project` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `service` varchar(255) NOT NULL,
  `namespace` varchar(255) DEFAULT NULL,
  `tenant_name` varchar(255) DEFAULT NULL,
  `sub_ids` varchar(255) DEFAULT NULL,
  `setting` text,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `UQE_project_name` (`name`),
  KEY `IDX_project_namespace` (`namespace`),
  KEY `IDX_project_service` (`service`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8;

-- 正在导出表  rtc.project 的数据：~12 rows (大约)
/*!40000 ALTER TABLE `project` DISABLE KEYS */;
INSERT INTO `project` (`id`, `name`, `service`, `namespace`, `tenant_name`, `sub_ids`, `setting`, `created_at`, `updated_at`) VALUES
	(1, 'go-api', 'go-api', '', 'test', 'null', '{"image":"xiaoxinmiao/go-api","envs":["APP_ENV=rtc"],"isProjectKafka":false,"ports":["8080"],"databases":{"mysql":["fruit"]},"streamNames":null}', '2019-09-28 12:43:12', '2019-09-28 12:43:12'),
	(2, 'go-api2', 'go-api2', '', 'pangpang', '[3,4]', '{"image":"registry.p2shop.com.cn/go-api","envs":["APP_ENV=rtc"],"isProjectKafka":false,"ports":["8080"],"databases":{"sqlserver":["fruit"]},"streamNames":["shipping","ipay"]}', '2019-09-28 12:43:12', '2019-09-28 12:43:12'),
	(3, 'go-api3', 'go-api3', '', 'pangpang', 'null', '{"image":"registry.p2shop.com.cn/go-api","envs":["APP_ENV=rtc"],"isProjectKafka":true,"ports":["8080"],"databases":{"redis":["fruit"]},"streamNames":["pangpangjan"]}', '2019-09-28 12:43:12', '2019-09-28 12:43:12'),
	(4, 'go-api4', 'go-api4', '', 'pangpang', 'null', '{"image":"registry.p2shop.com.cn/go-api","envs":["APP_ENV=rtc"],"isProjectKafka":true,"ports":["8080"],"databases":{"mysql":["fruit"]},"streamNames":["msl"]}', '2019-09-28 12:43:12', '2019-09-28 12:43:12'),
	(5, 'event-broker-kafka', 'event-broker-kafka', '', 'pangpang', 'null', '{"image":"registry.p2shop.com.cn/event-broker-kafka","envs":["KAFKA_BROKERS=rtc-kafka:9092"],"isProjectKafka":true,"ports":["3000"],"databases":null,"streamNames":null}', '2019-09-28 12:43:12', '2019-09-28 12:43:12'),
	(6, 'event-kafka-consumer', 'event-kafka-consumer', '', 'pangpang', 'null', '{"image":"registry.p2shop.com.cn/event-kafka-consumer","envs":["CONFIGOR_ENV=rtc"],"isProjectKafka":true,"ports":null,"databases":{"mysql":["event_broker"],"redis":null},"streamNames":null}', '2019-09-28 12:43:12', '2019-09-28 12:43:12'),
	(9, 'catalog', 'catalog', '', 'pangpang', 'null', '{"image":"registry.p2shop.com.cn/jan-catalog","envs":["POS_ENV=rtc","JWT_SECRET=******"],"isProjectKafka":false,"ports":["5000"],"databases":{"mysql":["catalog"]},"streamNames":["pangpangjan"]}', '2019-10-08 01:06:42', '2019-10-08 01:06:42'),
	(10, 'offer-api-pangpang-brand', 'offer-api', 'pangpang-brand', 'pangpang', 'null', '{"image":"registry.p2shop.com.cn/offer-api-pangpang-brand","envs":["APP_ENV=rtc"],"isProjectKafka":true,"ports":["5000"],"databases":{"mysql":["offer"],"redis":null},"streamNames":null}', '2019-10-09 01:06:42', '2019-10-09 01:06:42'),
	(11, 'product-api-pangpang-brand', 'product-api', 'pangpang-brand', 'pangpang', '[10]', '{"image":"registry.p2shop.com.cn/product-api-pangpang-brand","envs":["APP_ENV=rtc"],"isProjectKafka":true,"ports":["8000"],"databases":{"mysql":["product"]},"streamNames":null}', '2019-10-09 01:06:42', '2019-10-09 01:06:42'),
	(12, 'ipay-api', 'ipay-api', '', 'pangpang', 'null', '{"image":"registry.p2shop.com.cn/ipay-api","envs":["APP_ENV=rtc","IPAY_HOST=http://rtc-ipay-api:8080/v3","IPAY_CONN=root:1234@tcp(rtc-mysql:3306)/ipay?charset=utf8&parseTime=True&loc=UTC","JWT_SECRET=******","WX_NOTIFY_EVENT=http://rtc-event-broker-kafka:3000/v1/streams/ipay/events"],"isProjectKafka":true,"ports":["8080"],"databases":{"mysql":["ipay"]},"streamNames":["ipay"]}', '2019-10-09 01:06:42', '2019-10-09 01:06:42'),
	(17, 'go-api-test1', 'go-api-test1', '', 'pangpang', '[0]', '{"image":"registry.p2shop.com.cn/go-api-test1","envs":["APP_ENV=rtc"],"isProjectKafka":false,"ports":["8080"],"databases":{"mysql":["fruit"]},"streamNames":[""]}', '2019-10-12 02:50:01', '2019-10-12 02:50:01'),
	(18, 'rtc-api', 'rtc-api', '', 'pangpang', '', '{"image":"registry.p2shop.com.cn/rtc-api","envs":["APP_ENV=rtc"],"isProjectKafka":false,"ports":["8080"],"databases":{"mysql":["rtc"]},"streamNames":null}', '2019-10-12 02:50:01', '2019-10-12 02:50:01');
/*!40000 ALTER TABLE `project` ENABLE KEYS */;

-- 导出  表 rtc.tenant 结构
CREATE TABLE IF NOT EXISTS `tenant` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

-- 正在导出表  rtc.tenant 的数据：~3 rows (大约)
/*!40000 ALTER TABLE `tenant` DISABLE KEYS */;
INSERT INTO `tenant` (`id`, `name`) VALUES
	(1, 'pangpang'),
	(2, 'srx'),
	(3, 'test');
/*!40000 ALTER TABLE `tenant` ENABLE KEYS */;

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
