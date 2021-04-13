CREATE PROCEDURE `DropColumnIfExists` (`@TABLE` VARCHAR(100), `@COLUMN` VARCHAR(100))
`DropColumnIfExists`: BEGIN
    DECLARE `@EXISTS` INT UNSIGNED DEFAULT 0;

    SELECT COUNT(*) INTO `@EXISTS`
    FROM `information_schema`.`columns`
    WHERE (
                      `TABLE_SCHEMA` = DATABASE()
                  AND `TABLE_NAME` = `@TABLE`
                  AND `COLUMN_NAME` = `@COLUMN`
              );

    IF (`@EXISTS` > 0) THEN
        SET @SQL = CONCAT('ALTER TABLE `', `@TABLE`, '` DROP COLUMN `', `@COLUMN`, '`;');

        PREPARE query FROM @SQL;
        EXECUTE query;
    END IF;
END