-- MySQL dump 10.13  Distrib 5.7.34, for Linux (x86_64)
--
-- Host: localhost    Database: dhcpd
-- ------------------------------------------------------
-- Server version	5.7.34

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `acls`
--

DROP TABLE IF EXISTS `acls`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `acls` (
  `client_hw_addr` varchar(256) NOT NULL,
  `action` varchar(256) NOT NULL,
  PRIMARY KEY (`client_hw_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `acls`
--

LOCK TABLES `acls` WRITE;
/*!40000 ALTER TABLE `acls` DISABLE KEYS */;
/*!40000 ALTER TABLE `acls` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `bindings`
--

DROP TABLE IF EXISTS `bindings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `bindings` (
  `client_hw_addr` varchar(256) NOT NULL,
  `bind_addr` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`client_hw_addr`),
  UNIQUE KEY `bind_addr` (`bind_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `bindings`
--

LOCK TABLES `bindings` WRITE;
/*!40000 ALTER TABLE `bindings` DISABLE KEYS */;
/*!40000 ALTER TABLE `bindings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `leases`
--

DROP TABLE IF EXISTS `leases`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `leases` (
  `client_hw_addr` varchar(256) NOT NULL,
  `assigned_addr` varchar(256) DEFAULT NULL,
  `expires` datetime NOT NULL,
  PRIMARY KEY (`client_hw_addr`),
  UNIQUE KEY `assigned_addr` (`assigned_addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `leases`
--

LOCK TABLES `leases` WRITE;
/*!40000 ALTER TABLE `leases` DISABLE KEYS */;
/*!40000 ALTER TABLE `leases` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `options`
--

DROP TABLE IF EXISTS `options`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `options` (
  `lease_time` varchar(256) DEFAULT NULL,
  `server_ip` varchar(256) NOT NULL,
  `boot_file_name` varchar(256) DEFAULT NULL,
  `gateway_ip` varchar(256) DEFAULT NULL,
  `range_start_ip` varchar(256) DEFAULT NULL,
  `range_end_ip` varchar(256) DEFAULT NULL,
  `net_mask` varchar(256) DEFAULT NULL,
  `router` varchar(256) DEFAULT NULL,
  `dns` varchar(256) DEFAULT NULL,
  `acl` tinyint(1) DEFAULT NULL,
  `acl_action` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`server_ip`),
  UNIQUE KEY `lease_time` (`lease_time`),
  UNIQUE KEY `boot_file_name` (`boot_file_name`),
  UNIQUE KEY `gateway_ip` (`gateway_ip`),
  UNIQUE KEY `range_start_ip` (`range_start_ip`),
  UNIQUE KEY `range_end_ip` (`range_end_ip`),
  UNIQUE KEY `net_mask` (`net_mask`),
  UNIQUE KEY `router` (`router`),
  UNIQUE KEY `dns` (`dns`),
  UNIQUE KEY `acl` (`acl`),
  UNIQUE KEY `acl_action` (`acl_action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `options`
--

LOCK TABLES `options` WRITE;
/*!40000 ALTER TABLE `options` DISABLE KEYS */;
INSERT INTO `options` VALUES ('1h','127.0.0.1','pxelinux.0','10.1.1.1','10.1.1.10','10.1.1.100','255.0.0.0','10.1.1.1','10.1.1.1',1,'');
/*!40000 ALTER TABLE `options` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `reserves`
--

DROP TABLE IF EXISTS `reserves`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `reserves` (
  `address` varchar(256) NOT NULL,
  PRIMARY KEY (`address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `reserves`
--

LOCK TABLES `reserves` WRITE;
/*!40000 ALTER TABLE `reserves` DISABLE KEYS */;
/*!40000 ALTER TABLE `reserves` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-06-26 22:08:43
