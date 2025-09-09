-- MySQL dump 10.13  Distrib 8.0.43, for Linux (aarch64)
--
-- Host: localhost    Database: atm
-- ------------------------------------------------------
-- Server version	8.0.43

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `caja`
--

DROP TABLE IF EXISTS `caja`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `caja` (
  `id` int NOT NULL AUTO_INCREMENT,
  `saldo` decimal(15,2) NOT NULL DEFAULT '0.00',
  `ultima_actualizacion` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `caja`
--

LOCK TABLES `caja` WRITE;
/*!40000 ALTER TABLE `caja` DISABLE KEYS */;
INSERT INTO `caja` VALUES (1,650000.00,'2025-09-09 14:33:16');
/*!40000 ALTER TABLE `caja` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `cajas`
--

DROP TABLE IF EXISTS `cajas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cajas` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `saldo` decimal(15,2) NOT NULL DEFAULT '0.00',
  `ultima_actualizacion` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `cajas`
--

LOCK TABLES `cajas` WRITE;
/*!40000 ALTER TABLE `cajas` DISABLE KEYS */;
/*!40000 ALTER TABLE `cajas` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `categoria`
--

DROP TABLE IF EXISTS `categoria`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categoria` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nombre` varchar(100) NOT NULL,
  `tipo` varchar(20) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `categoria`
--

LOCK TABLES `categoria` WRITE;
/*!40000 ALTER TABLE `categoria` DISABLE KEYS */;
/*!40000 ALTER TABLE `categoria` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `categorias`
--

DROP TABLE IF EXISTS `categorias`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `categorias` (
  `id` int NOT NULL AUTO_INCREMENT,
  `nombre` varchar(100) NOT NULL,
  `tipo` enum('INGRESO','EGRESO') NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `categorias`
--

LOCK TABLES `categorias` WRITE;
/*!40000 ALTER TABLE `categorias` DISABLE KEYS */;
INSERT INTO `categorias` VALUES (1,'Cartera de Clientes','INGRESO'),(2,'Efectivo Puntos de Venta','INGRESO'),(3,'Logística y Transporte','INGRESO'),(4,'Trading','INGRESO'),(5,'Operación Monetaria','INGRESO'),(6,'Retiros Bancarios','INGRESO'),(7,'Otros Ingresos','INGRESO'),(8,'Pago a Proveedores','EGRESO'),(9,'Gastos Operativos','EGRESO'),(10,'Logística y Transporte','EGRESO'),(11,'Nómina y Beneficios','EGRESO'),(12,'Obligaciones Bancarias','EGRESO'),(13,'Operación Monetaria','EGRESO'),(14,'Otros Egresos','EGRESO');
/*!40000 ALTER TABLE `categorias` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Temporary view structure for view `resumen_financiero`
--

DROP TABLE IF EXISTS `resumen_financiero`;
/*!50001 DROP VIEW IF EXISTS `resumen_financiero`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `resumen_financiero` AS SELECT 
 1 AS `total_ingresos`,
 1 AS `total_egresos`,
 1 AS `balance_general`,
 1 AS `saldo_en_caja`*/;
SET character_set_client = @saved_cs_client;

--
-- Table structure for table `transaccion_logs`
--

DROP TABLE IF EXISTS `transaccion_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `transaccion_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `transaccion_id` bigint unsigned DEFAULT NULL,
  `accion` varchar(10) NOT NULL,
  `monto` decimal(15,2) DEFAULT NULL,
  `descripcion` longtext,
  `fecha` datetime(3) DEFAULT NULL,
  `usuario` longtext,
  `saldo_antes` decimal(15,2) DEFAULT NULL,
  `saldo_despues` decimal(15,2) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `transaccion_logs`
--

LOCK TABLES `transaccion_logs` WRITE;
/*!40000 ALTER TABLE `transaccion_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `transaccion_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `transacciones`
--

DROP TABLE IF EXISTS `transacciones`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `transacciones` (
  `id` int NOT NULL AUTO_INCREMENT,
  `categoria_id` int NOT NULL,
  `monto` decimal(15,2) NOT NULL,
  `fecha` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `descripcion` text NOT NULL,
  PRIMARY KEY (`id`),
  KEY `categoria_id` (`categoria_id`),
  CONSTRAINT `transacciones_ibfk_1` FOREIGN KEY (`categoria_id`) REFERENCES `categorias` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `transacciones`
--

LOCK TABLES `transacciones` WRITE;
/*!40000 ALTER TABLE `transacciones` DISABLE KEYS */;
INSERT INTO `transacciones` VALUES (1,2,1000000.00,'2025-09-08 13:04:50','EFECTIVO BURBUJA'),(2,2,1000000.00,'2025-09-08 13:08:22','EFECTIVO VISTO'),(3,11,1800000.00,'2025-09-08 13:09:38','NOMINA BODEGA 2'),(5,2,200000.00,'2025-09-08 19:16:09','EFECTIVO VISTO'),(6,3,150000.00,'2025-09-08 19:19:40','PAGO BOYACO'),(9,1,100000.00,'2025-09-08 21:47:51','ABONO AUGUSTO'),(11,5,20000000.00,'2025-09-08 21:54:14','ABONO A FACTURA AUGUSTO'),(12,10,15000000.00,'2025-09-08 22:03:40','PAGO CAMIONES'),(13,14,5000000.00,'2025-09-09 08:11:41','PAGO TARJETA DANIELA');
/*!40000 ALTER TABLE `transacciones` ENABLE KEYS */;
UNLOCK TABLES;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_actualizar_caja_insert` AFTER INSERT ON `transacciones` FOR EACH ROW BEGIN
    DECLARE tipo_categoria ENUM('INGRESO','EGRESO');
    SELECT tipo INTO tipo_categoria FROM categorias WHERE id = NEW.categoria_id;

    IF tipo_categoria = 'INGRESO' THEN
        UPDATE caja SET saldo = saldo + NEW.monto, ultima_actualizacion = NOW() WHERE id = 1;
    ELSE
        UPDATE caja SET saldo = saldo - NEW.monto, ultima_actualizacion = NOW() WHERE id = 1;
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_log_insert` AFTER INSERT ON `transacciones` FOR EACH ROW BEGIN
    DECLARE tipo_categoria ENUM('INGRESO','EGRESO');
    DECLARE saldo_anterior DECIMAL(15,2);
    DECLARE saldo_despues DECIMAL(15,2);

    -- saldo actual de la caja después de aplicar la transacción
    SELECT saldo INTO saldo_despues FROM caja WHERE id = 1;

    -- obtener tipo de categoría
    SELECT tipo INTO tipo_categoria FROM categorias WHERE id = NEW.categoria_id;

    -- calcular saldo antes en base al saldo_despues
    IF tipo_categoria = 'INGRESO' THEN
        SET saldo_anterior = saldo_despues - NEW.monto;
    ELSE
        SET saldo_anterior = saldo_despues + NEW.monto;
    END IF;

    INSERT INTO transacciones_log (transaccion_id, accion, monto, descripcion, usuario, saldo_antes, saldo_despues)
    VALUES (NEW.id, 'INSERT', NEW.monto, NEW.descripcion, CURRENT_USER(), saldo_anterior, saldo_despues);
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_actualizar_caja_update` AFTER UPDATE ON `transacciones` FOR EACH ROW BEGIN
    DECLARE tipo_categoria ENUM('INGRESO','EGRESO');
    SELECT tipo INTO tipo_categoria FROM categorias WHERE id = NEW.categoria_id;

    IF tipo_categoria = 'INGRESO' THEN
        UPDATE caja SET saldo = saldo - OLD.monto + NEW.monto, ultima_actualizacion = NOW() WHERE id = 1;
    ELSE
        UPDATE caja SET saldo = saldo + OLD.monto - NEW.monto, ultima_actualizacion = NOW() WHERE id = 1;
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_log_update` AFTER UPDATE ON `transacciones` FOR EACH ROW BEGIN
    DECLARE tipo_categoria ENUM('INGRESO','EGRESO');
    DECLARE saldo_anterior DECIMAL(15,2);
    DECLARE saldo_despues DECIMAL(15,2);

    -- saldo ya actualizado en caja
    SELECT saldo INTO saldo_despues FROM caja WHERE id = 1;

    -- tipo de categoría
    SELECT tipo INTO tipo_categoria FROM categorias WHERE id = NEW.categoria_id;

    -- calcular saldo antes con base en la diferencia aplicada
    IF tipo_categoria = 'INGRESO' THEN
        SET saldo_anterior = saldo_despues - NEW.monto + OLD.monto;
    ELSE
        SET saldo_anterior = saldo_despues + NEW.monto - OLD.monto;
    END IF;

    INSERT INTO transacciones_log (transaccion_id, accion, monto, descripcion, usuario, saldo_antes, saldo_despues)
    VALUES (NEW.id, 'UPDATE', NEW.monto, NEW.descripcion, CURRENT_USER(), saldo_anterior, saldo_despues);
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_actualizar_caja_delete` AFTER DELETE ON `transacciones` FOR EACH ROW BEGIN
    DECLARE tipo_categoria ENUM('INGRESO','EGRESO');
    SELECT tipo INTO tipo_categoria FROM categorias WHERE id = OLD.categoria_id;

    IF tipo_categoria = 'INGRESO' THEN
        UPDATE caja SET saldo = saldo - OLD.monto, ultima_actualizacion = NOW() WHERE id = 1;
    ELSE
        UPDATE caja SET saldo = saldo + OLD.monto, ultima_actualizacion = NOW() WHERE id = 1;
    END IF;
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `trg_log_delete` AFTER DELETE ON `transacciones` FOR EACH ROW BEGIN
    DECLARE tipo_categoria ENUM('INGRESO','EGRESO');
    DECLARE saldo_anterior DECIMAL(15,2);
    DECLARE saldo_despues DECIMAL(15,2);

    SELECT saldo INTO saldo_despues FROM caja WHERE id = 1;
    SELECT tipo INTO tipo_categoria FROM categorias WHERE id = OLD.categoria_id;

    IF tipo_categoria = 'INGRESO' THEN
        SET saldo_anterior = saldo_despues + OLD.monto;
    ELSE
        SET saldo_anterior = saldo_despues - OLD.monto;
    END IF;

    INSERT INTO transacciones_log (transaccion_id, accion, monto, descripcion, usuario, saldo_antes, saldo_despues)
    VALUES (OLD.id, 'DELETE', OLD.monto, OLD.descripcion, CURRENT_USER(), saldo_anterior, saldo_despues);
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `transacciones_log`
--

DROP TABLE IF EXISTS `transacciones_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `transacciones_log` (
  `id` int NOT NULL AUTO_INCREMENT,
  `transaccion_id` int DEFAULT NULL,
  `accion` enum('INSERT','UPDATE','DELETE') NOT NULL,
  `monto` decimal(15,2) DEFAULT NULL,
  `descripcion` text,
  `fecha` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `usuario` varchar(100) DEFAULT NULL,
  `saldo_antes` decimal(15,2) DEFAULT NULL,
  `saldo_despues` decimal(15,2) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `transacciones_log`
--

LOCK TABLES `transacciones_log` WRITE;
/*!40000 ALTER TABLE `transacciones_log` DISABLE KEYS */;
INSERT INTO `transacciones_log` VALUES (1,1,'INSERT',1000000.00,'EFECTIVO BURBUJA','2025-09-08 18:04:50','root@%',1000000.00,2000000.00),(2,2,'INSERT',1000000.00,'EFECTIVO VISTO','2025-09-08 18:08:21','root@%',1000000.00,2000000.00),(3,3,'INSERT',2000000.00,'NOMINA BODEGA','2025-09-08 18:09:37','root@%',2000000.00,0.00),(4,3,'UPDATE',2000000.00,'NOMINA BODEGA 2','2025-09-08 18:17:56','root@%',0.00,0.00),(5,3,'UPDATE',1800000.00,'NOMINA BODEGA 2','2025-09-08 18:18:30','root@%',400000.00,200000.00),(6,4,'INSERT',500000.00,'EFECTIVO MADRUGON GRAN SAN','2025-09-08 19:11:40','root@%',200000.00,700000.00),(7,4,'UPDATE',450000.00,'EFECTIVO MADRUGON GRAN SAN','2025-09-08 19:12:23','root@%',700000.00,650000.00),(8,4,'DELETE',450000.00,'EFECTIVO MADRUGON GRAN SAN','2025-09-08 19:13:07','root@%',650000.00,200000.00),(9,5,'INSERT',200000.00,'EFECTIVO VISTO','2025-09-09 00:16:08','root@%',200000.00,400000.00),(10,6,'INSERT',150000.00,'PAGO BOYACO','2025-09-09 00:19:39','root@%',400000.00,550000.00),(11,7,'INSERT',100000.00,'PAGO TARJETA','2025-09-09 00:30:06','root@%',550000.00,450000.00),(12,7,'UPDATE',200000.00,'PAGO TARJETA','2025-09-09 00:31:29','root@%',450000.00,350000.00),(13,7,'DELETE',200000.00,'PAGO TARJETA','2025-09-09 00:31:45','root@%',350000.00,550000.00),(14,8,'INSERT',100000.00,'PAGO TARJETA','2025-09-09 00:42:29','root@%',550000.00,450000.00),(15,8,'UPDATE',200000.00,'PAGO TARJETA','2025-09-09 00:42:53','root@%',450000.00,350000.00),(16,8,'UPDATE',200000.00,'PAGO CUOTA','2025-09-09 00:43:29','root@%',350000.00,350000.00),(17,8,'DELETE',200000.00,'PAGO CUOTA','2025-09-09 00:43:40','root@%',350000.00,550000.00),(18,9,'INSERT',100000.00,'ABONO AUGUSTO','2025-09-09 02:47:50','root@%',550000.00,650000.00),(19,10,'INSERT',10000.00,'PRUEBA DE EGRESO','2025-09-09 02:52:47','root@%',650000.00,640000.00),(20,11,'INSERT',10000000.00,'ABONO A FACTURA','2025-09-09 02:54:14','root@%',640000.00,10640000.00),(21,11,'UPDATE',10000000.00,'ABONO A FACTURA AUGUSTO','2025-09-09 03:01:39','root@%',10640000.00,10640000.00),(22,11,'UPDATE',10000000.00,'ABONO A FACTURA AUGUSTO','2025-09-09 03:02:08','root@%',10640000.00,10640000.00),(23,11,'UPDATE',20000000.00,'ABONO A FACTURA AUGUSTO','2025-09-09 03:02:41','root@%',10640000.00,20640000.00),(24,10,'DELETE',10000.00,'PRUEBA DE EGRESO','2025-09-09 03:03:01','root@%',20640000.00,20650000.00),(25,12,'INSERT',15000000.00,'PAGO CAMIONES','2025-09-09 03:03:40','root@%',20650000.00,5650000.00),(26,13,'INSERT',5000000.00,'PAGO TARJETA DANIELA','2025-09-09 13:11:40','root@%',5650000.00,650000.00),(27,13,'UPDATE',5000000.00,'PAGO TARJETA DANIELA','2025-09-09 14:33:16','root@%',650000.00,650000.00);
/*!40000 ALTER TABLE `transacciones_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `transaccions`
--

DROP TABLE IF EXISTS `transaccions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `transaccions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `categoria_id` bigint unsigned NOT NULL,
  `monto` decimal(15,2) NOT NULL,
  `fecha` datetime(3) DEFAULT NULL,
  `descripcion` text NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_categoria_transacciones` (`categoria_id`),
  CONSTRAINT `fk_categoria_transacciones` FOREIGN KEY (`categoria_id`) REFERENCES `categoria` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `transaccions`
--

LOCK TABLES `transaccions` WRITE;
/*!40000 ALTER TABLE `transaccions` DISABLE KEYS */;
/*!40000 ALTER TABLE `transaccions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Final view structure for view `resumen_financiero`
--

/*!50001 DROP VIEW IF EXISTS `resumen_financiero`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_0900_ai_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`root`@`%` SQL SECURITY DEFINER */
/*!50001 VIEW `resumen_financiero` AS select (select ifnull(sum(`t`.`monto`),0) from (`transacciones` `t` join `categorias` `c` on((`c`.`id` = `t`.`categoria_id`))) where (`c`.`tipo` = 'INGRESO')) AS `total_ingresos`,(select ifnull(sum(`t`.`monto`),0) from (`transacciones` `t` join `categorias` `c` on((`c`.`id` = `t`.`categoria_id`))) where (`c`.`tipo` = 'EGRESO')) AS `total_egresos`,((select ifnull(sum(`t`.`monto`),0) from (`transacciones` `t` join `categorias` `c` on((`c`.`id` = `t`.`categoria_id`))) where (`c`.`tipo` = 'INGRESO')) - (select ifnull(sum(`t`.`monto`),0) from (`transacciones` `t` join `categorias` `c` on((`c`.`id` = `t`.`categoria_id`))) where (`c`.`tipo` = 'EGRESO'))) AS `balance_general`,(select `caja`.`saldo` from `caja` where (`caja`.`id` = 1)) AS `saldo_en_caja` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-09-09 15:11:59
