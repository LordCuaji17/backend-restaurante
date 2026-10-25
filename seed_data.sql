/*M!999999\- enable the sandbox mode */ 
-- MariaDB dump 10.19-11.8.6-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: db_restaurante
-- ------------------------------------------------------
-- Server version	11.8.6-MariaDB-5ubuntu0.1 from Ubuntu

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*M!100616 SET @OLD_NOTE_VERBOSITY=@@NOTE_VERBOSITY, NOTE_VERBOSITY=0 */;

--
-- Table structure for table `mesas`
--

DROP TABLE IF EXISTS `mesas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `mesas` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `numero_mesa` int(11) NOT NULL,
  `estado` enum('Libre','Ocupada') DEFAULT 'Libre',
  PRIMARY KEY (`id`),
  UNIQUE KEY `numero_mesa` (`numero_mesa`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mesas`
--

SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT, @@AUTOCOMMIT=0;
LOCK TABLES `mesas` WRITE;
/*!40000 ALTER TABLE `mesas` DISABLE KEYS */;
INSERT INTO `mesas` VALUES
(1,1,'Ocupada'),
(2,2,'Libre'),
(3,3,'Libre'),
(4,4,'Libre');
/*!40000 ALTER TABLE `mesas` ENABLE KEYS */;
UNLOCK TABLES;
COMMIT;
SET AUTOCOMMIT=@OLD_AUTOCOMMIT;

--
-- Table structure for table `productos`
--

DROP TABLE IF EXISTS `productos`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `productos` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `nombre` varchar(100) NOT NULL,
  `precio` decimal(10,2) NOT NULL,
  `categoria` varchar(50) NOT NULL,
  `activo` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=56 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `productos`
--

SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT, @@AUTOCOMMIT=0;
LOCK TABLES `productos` WRITE;
/*!40000 ALTER TABLE `productos` DISABLE KEYS */;
INSERT INTO `productos` VALUES
(1,'Piña Rellena de Mariscos',185.00,'Especialidades',1),
(2,'Plátano Relleno de Mariscos',185.00,'Especialidades',1),
(3,'Cazuela Jarocha',150.00,'Especialidades',1),
(4,'Arroz a la Tumbada',180.00,'Especialidades',1),
(5,'Mixiote de Mariscos',150.00,'Especialidades',1),
(6,'Alambre de Mariscos',150.00,'Especialidades',1),
(7,'Aguachile de Camarón',185.00,'Especialidades',1),
(8,'Chile Relleno Camarón y Pulpo',165.00,'Especialidades',1),
(9,'Chile Relleno de Camarón',165.00,'Especialidades',1),
(10,'Chile Relleno de Mariscos',165.00,'Especialidades',1),
(11,'Camarones Empanizados',170.00,'Camarones',1),
(12,'Camarones Al Mojo de Ajo',170.00,'Camarones',1),
(13,'Camarones Al Ajillo',170.00,'Camarones',1),
(14,'Camarones A la Diabla',170.00,'Camarones',1),
(15,'Camarones Al Chipotle',170.00,'Camarones',1),
(16,'Camarones Al Coco',170.00,'Camarones',1),
(17,'Filete Chicharrón de Pescado',160.00,'Filetes',1),
(18,'Filete Empanizado',130.00,'Filetes',1),
(19,'Filete A la Parmesana',140.00,'Filetes',1),
(20,'Filete A la Diabla',140.00,'Filetes',1),
(21,'Filete A la Plancha',130.00,'Filetes',1),
(22,'Filete Relleno de Mariscos',185.00,'Filetes',1),
(23,'Empanada de Atún',40.00,'Empanadas',1),
(24,'Empanada de Camarón',50.00,'Empanadas',1),
(25,'Empanada de Camarón con Pulpo',60.00,'Empanadas',1),
(26,'Orden Pescadillas (2)',50.00,'Quesadillas',1),
(27,'Coctel Camarón Chico',90.00,'Cocteles',1),
(28,'Coctel Pulpo Chico',90.00,'Cocteles',1),
(29,'Coctel Campechano Chico',90.00,'Cocteles',1),
(30,'Coctel Camarón Grande',165.00,'Cocteles',1),
(31,'Coctel Pulpo Grande',165.00,'Cocteles',1),
(32,'Coctel Campechano Grande',165.00,'Cocteles',1),
(33,'Coctel Camarón En bola',200.00,'Cocteles',1),
(34,'Coctel Campechano En bola',200.00,'Cocteles',1),
(35,'Caldo de Camarón Pelado',165.00,'Caldos',1),
(36,'Caldo de Camarón, Pescado',140.00,'Caldos',1),
(37,'Sopa de Mariscos',165.00,'Caldos',1),
(38,'Caldo de Camarón (para pelar)',140.00,'Caldos',1),
(39,'Caldo de Pescado',130.00,'Caldos',1),
(40,'Mojarra Frita',150.00,'Mojarras',1),
(41,'Mojarra Al mojo de Ajo',150.00,'Mojarras',1),
(42,'Mojarra A la Diabla',150.00,'Mojarras',1),
(43,'Mojarra Empapelada',150.00,'Mojarras',1),
(44,'Orden de Arroz',50.00,'Órdenes Extra',1),
(45,'Orden de Arroz con Queso',55.00,'Órdenes Extra',1),
(46,'Orden de Arroz con Camarones',90.00,'Órdenes Extra',1),
(47,'Orden de Ensalada',40.00,'Órdenes Extra',1),
(48,'Plátanos con Crema',50.00,'Postres',1),
(49,'Duraznos con Rompope',50.00,'Postres',1),
(50,'Ate con Queso',50.00,'Postres',1),
(51,'Refresco 355 ml',30.00,'Bebidas',1),
(52,'Refresco 600 ml',35.00,'Bebidas',1),
(53,'Agua 500 ml',30.00,'Bebidas',1),
(55,'Cerveza',45.00,'Personalizados',1);
/*!40000 ALTER TABLE `productos` ENABLE KEYS */;
UNLOCK TABLES;
COMMIT;
SET AUTOCOMMIT=@OLD_AUTOCOMMIT;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*M!100616 SET NOTE_VERBOSITY=@OLD_NOTE_VERBOSITY */;

-- Dump completed on 2026-10-25 12:36:00
