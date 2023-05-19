SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
    `username` varchar(255) PRIMARY KEY,
    `id` varchar(255) NOT NULL,
    `password` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `users` (`username`, `id`, `password`) VALUES
('admin1', 'id_admin1', 'adminadmin'),
('admin2', 'id_admin2', 'adminadmin');