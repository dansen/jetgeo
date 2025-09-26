package com.ling5821.jetgeo.server;

import com.ling5821.jetgeo.JetGeo;
import com.ling5821.jetgeo.model.GeoInfo;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequiredArgsConstructor
public class ReverseGeoController {

    private final JetGeo jetGeo;

    @GetMapping("/api/reverse")
    public ResponseEntity<?> reverse(@RequestParam("lat") double lat,
                                     @RequestParam("lng") double lng) {
        GeoInfo info = jetGeo.getGeoInfo(lat, lng);
        if (info == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(info);
    }
}
