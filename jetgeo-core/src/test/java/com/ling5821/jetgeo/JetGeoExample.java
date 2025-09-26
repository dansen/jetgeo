package com.ling5821.jetgeo;

import com.ling5821.jetgeo.config.JetGeoProperties;
import com.ling5821.jetgeo.enums.LevelEnum;
import com.ling5821.jetgeo.model.GeoInfo;

/**
 * @author lsj
 * @date 2021/11/20 14:34
 */
public class JetGeoExample {

    public static final JetGeo jetGeo;

    static {
        JetGeoProperties properties = new JetGeoProperties();
        // Resolve geo data path: prefer environment variable GEO_DATA_PATH, else fallback to ./data/geodata
        String envPath = System.getenv("GEO_DATA_PATH");
        if (envPath == null || envPath.trim().isEmpty()) {
            // relative to project root when running via maven (user.dir)
            String userDir = System.getProperty("user.dir");
            envPath = userDir + java.io.File.separator + "data" + java.io.File.separator + "geodata";
        }
        properties.setGeoDataParentPath(envPath);
        // Set finest level you want to load (district loads all)
        properties.setLevel(LevelEnum.district);
        jetGeo = new JetGeo(properties);
        System.out.println("[JetGeoExample] Using GEO_DATA_PATH=" + envPath);
    }

    public static void main(String[] args) {
        GeoInfo geoInfo = jetGeo.getGeoInfo(32.053197915979325, 118.85999259252777);
        System.out.println(geoInfo);
    }
}
